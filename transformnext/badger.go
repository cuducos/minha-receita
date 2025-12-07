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

// As of 2025-11 the longest sequence we've got was 257, so setting it to 512 to
// have some room — maybe this could be set from a CLI flag to avoid recompiling
// when source data changes and needs more space.
const defaultPoolSize = 512

type kv struct {
	db   *badger.DB
	pool sync.Pool
}

func (kv *kv) serialize(b []byte, row []string) ([]byte, error) {
	var err error
	for _, v := range row {
		s := uint32(len(v)) // used to deserialize later on
		b, err = binary.Append(b, binary.LittleEndian, s)
		if err != nil {
			return nil, err
		}
		b = append(b, v...)
	}
	return b, nil
}

func (kv *kv) deserialize(val []byte) ([]string, error) {
	if val == nil {
		return nil, nil
	}
	var out []string
	r := bytes.NewReader(val)
	for r.Len() > 0 {
		err := func() error {
			var s uint32
			if err := binary.Read(r, binary.LittleEndian, &s); err != nil {
				return fmt.Errorf("error reading size: %w", err)
			}
			b := kv.pool.Get().(*[]byte)
			*b = (*b)[:s]
			defer kv.pool.Put(b)
			if cap(*b) < int(s) {
				return fmt.Errorf("buffer from pool too small (%d): needs %d", cap(*b), s)
			}
			n, err := io.ReadFull(r, *b)
			if err != nil {
				return fmt.Errorf("could not deserialize value: %w", err)
			}
			if n != int(s) {
				return fmt.Errorf("expected to read %d bytes, got %d", s, n)
			}
			out = append(out, string(*b))
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (kv *kv) put(src *source, id string, row []string) error {
	if len(row) == 0 {
		return nil
	}
	key := src.keyFor(id)
	b := kv.pool.Get().(*[]byte)
	*b = (*b)[:0]
	defer kv.pool.Put(b)
	val, err := kv.serialize(*b, row)
	if err != nil {
		return fmt.Errorf("could not serialize row %v: %w", row, err)
	}
	return kv.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (kv *kv) get(k []byte) ([]string, error) {
	val := kv.pool.Get().(*[]byte)
	*val = (*val)[:0]
	defer kv.pool.Put(val)
	err := kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return fmt.Errorf("could not get key: %w", err)
		}
		*val, err = item.ValueCopy(*val)
		if err != nil {
			return fmt.Errorf("could not read value: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not get key %s: %w", string(k), err)
	}
	return kv.deserialize(*val)
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
	if os.Getenv("DEBUG") != "badger" { // TODO: remove that after moving transformnext into transform
		opt = opt.WithLogger(&noLogger{})
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, fmt.Errorf("could not open badger at %s: %w", dir, err)
	}
	kv := &kv{
		db: db,
		pool: sync.Pool{
			New: func() any {
				b := make([]byte, defaultPoolSize)
				return &b
			},
		},
	}
	return kv, nil
}
