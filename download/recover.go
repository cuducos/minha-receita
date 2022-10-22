package download

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const recoverFileName = ".downloading"

type recover struct {
	dir       string
	chunkSize int
	files     map[string][]bool
	mutex     sync.Mutex
}

func (r *recover) path() string { return filepath.Join(r.dir, recoverFileName) }

func (r *recover) shouldDownload(f string, idx int) bool {
	f = filepath.Base(f)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return !r.files[f][idx]
}

func (r *recover) save() error {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("%d\n", r.chunkSize))
	r.mutex.Lock()
	for f, cs := range r.files {
		b.WriteString(fmt.Sprintf("%s:", f))
		for _, s := range cs {
			if s {
				b.WriteString("1")
			} else {
				b.WriteString("0")
			}
		}
		b.WriteString("\n")
	}
	r.mutex.Unlock()
	if err := os.WriteFile(r.path(), b.Bytes(), 0755); err != nil {
		return fmt.Errorf("error writing to %s: %w", r.path(), err)
	}
	return nil
}

func (r *recover) load(restart bool) error {
	if restart {
		err := os.Remove(r.path())
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not delete %s: %w", r.path(), err)
		}
		return nil
	}

	f, err := os.Open(r.path())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error opening %s: %w", r.path(), err)
	}
	defer f.Close()

	r.mutex.Lock()
	s := bufio.NewScanner(f)
	var i int
	for s.Scan() {
		l := s.Text()
		if i == 0 {
			n, err := strconv.Atoi(l)
			if err != nil {
				return fmt.Errorf("could not convert chunk size %s to number: %w", l, err)
			}
			if n != r.chunkSize {
				return fmt.Errorf(
					"chunk size in %s is %d, but it is %d for the current download; use --restart to start from scratch or run the download with --chunk-size=%d",
					r.path(),
					n,
					r.chunkSize,
					n,
				)
			}
		} else {
			p := strings.Split(l, ":")
			r.files[p[0]] = []bool{}
			for _, n := range p[1] {
				r.files[p[0]] = append(r.files[p[0]], n == '1')
			}
		}
		i += 1
	}
	r.mutex.Unlock()
	return nil
}

func (r *recover) addFile(f string, c int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	f = filepath.Base(f)
	_, ok := r.files[f]
	if ok {
		return
	}
	r.files[f] = make([]bool, c)
}

func (r *recover) chunkDone(f string, idx int) {
	f = filepath.Base(f)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.files[f][idx] = true
}

func (r *recover) isDone(f string) bool {
	f = filepath.Base(f)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	v, ok := r.files[f]
	if !ok {
		return false
	}
	for _, d := range v {
		if !d {
			return false
		}
	}
	return true
}

func (r *recover) close() error {
	for f := range r.files {
		if !r.isDone(f) {
			return nil
		}
	}
	err := os.Remove(r.path())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error cleaning up recovery file %s: %w", r.path(), err)
	}
	return nil
}

func newRecover(dir string, s int, x bool) (*recover, error) {
	r := recover{dir: dir, chunkSize: s, files: make(map[string][]bool)}
	if err := r.load(x); err != nil {
		return nil, fmt.Errorf("error loading %s: %w", r.path(), err)
	}
	return &r, nil
}
