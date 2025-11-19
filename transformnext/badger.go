package transformnext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
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
	for r.Len() > 0 {
		var s uint32
		if err := binary.Read(r, binary.LittleEndian, &s); err != nil {
			return nil, fmt.Errorf("error reading size: %w", err)
		}
		raw := kv.bytes.Get().(*[]byte)
		if cap(*raw) < int(s) {
			return nil, fmt.Errorf("buffer from pool too small (%d): needs %d", cap(*raw), s)
		} else {
			*raw = (*raw)[:s]
		}
		n, err := io.ReadFull(r, *raw)
		if err != nil {
			kv.bytes.Put(raw)
			return nil, fmt.Errorf("could not deserialize value: %w", err)
		}
		if n != int(s) {
			kv.bytes.Put(raw)
			return nil, fmt.Errorf("expected to read %d bytes, got %d", s, n)
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

func (kv *kv) getPrefix(k []byte) ([][]string, error) {
	vs := [][]string{}
	err := kv.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(k); it.ValidForPrefix(k); it.Next() {
			i := it.Item()
			err := i.Value(func(b []byte) error {
				v, err := kv.deserialize(b)
				if err != nil {
					return fmt.Errorf("could not deserialize %s: %w", string(i.Key()), err)
				}
				vs = append(vs, v)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vs, nil
}

type noLogger struct{}

func (*noLogger) Errorf(string, ...any)   {}
func (*noLogger) Warningf(string, ...any) {}
func (*noLogger) Infof(string, ...any)    {}
func (*noLogger) Debugf(string, ...any)   {}

func newBadger(dir string, ro bool) (*kv, error) {
	opt := badger.DefaultOptions(dir).WithReadOnly(ro).WithBypassLockGuard(true).WithDetectConflicts(false)
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
			// as of 2025-11 the longest sequence we've got was 159, so setting
			// it to 256 to have some room — but this could be set from a cli
			// flag to avoid recompiling when source data changes and needs
			// more space
			b := make([]byte, 0, 256)
			return &b
		},
	}
	return kv, nil
}
