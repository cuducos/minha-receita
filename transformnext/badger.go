package transformnext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

type kv struct {
	db    *badger.DB
	buf   sync.Pool
	bytes sync.Pool
}

func (kv *kv) serialize(row []string) ([]byte, error) {
	buf := kv.buf.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		kv.buf.Put(buf)
	}()
	for _, v := range row {
		s := uint32(len(v)) // used to deserialize later on
		if err := binary.Write(buf, binary.LittleEndian, s); err != nil {
			return nil, err
		}
		if _, err := buf.Write([]byte(v)); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (kv *kv) deserialize(b []byte) ([]string, error) {
	if b == nil {
		return nil, nil
	}
	var out []string
	r := bytes.NewReader(b)
	for {
		var s uint32
		if err := binary.Read(r, binary.LittleEndian, &s); err != nil {
			break
		}
		raw := kv.bytes.Get().(*[]byte)
		if cap(*raw) < int(s) {
			*raw = make([]byte, s)
		} else {
			*raw = (*raw)[:s]
		}
		if _, err := r.Read(*raw); err != nil {
			kv.bytes.Put(raw)
			return nil, fmt.Errorf("could not deserialize value: %w", err)
		}
		out = append(out, string(*raw))
		kv.bytes.Put(raw)
	}
	return out, nil
}

func (kv *kv) put(src *source, id string, row []string) error {
	if len(row) == 0 {
		return nil
	}
	k := src.keyFor(id)
	v, err := kv.serialize(row)
	if err != nil {
		return fmt.Errorf("could not serialize row %v: %w", row, err)
	}
	return kv.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
}

func (kv *kv) get(k []byte) ([]string, error) {
	var b []byte
	err := kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("could not get key: %w", err)
		}
		b, err = item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("could not read value: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not get key %s: %w", string(k), err)
	}
	return kv.deserialize(b)
}

type noLogger struct{}

func (*noLogger) Errorf(string, ...any)   {}
func (*noLogger) Warningf(string, ...any) {}
func (*noLogger) Infof(string, ...any)    {}
func (*noLogger) Debugf(string, ...any)   {}

func newBadger(dir string, ro bool) (*kv, error) {
	opt := badger.DefaultOptions(dir).WithReadOnly(ro)
	slog.Debug("Creating temporary key-value storage", "path", dir)
	if os.Getenv("DEBUG") == "" {
		opt = opt.WithLogger(&noLogger{})
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, fmt.Errorf("could not open badger at %s: %w", dir, err)
	}
	kv := &kv{db: db}
	kv.buf = sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
	kv.bytes = sync.Pool{
		New: func() any {
			b := make([]byte, 0, 1024)
			return &b
		},
	}
	return kv, nil
}
