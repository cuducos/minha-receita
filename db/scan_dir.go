package db

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cuducos/minha-receita/transform"
)

const maxReadDirEntries = 128

func isValidJSON(pth string) bool {
	if !strings.HasSuffix(pth, ".json") {
		return false
	}
	_, err := transform.CNPJForPath(pth)
	if err != nil {
		return false
	}
	return true
}

// isDir wraps a open/close of a path descriptor, thus each readDir gorotine can
// have only one descriptor opened at a time
func isDir(pth string) (bool, error) {
	f, err := os.Open(pth)
	if err != nil {
		if os.IsPermission(err) { // skip directories we don't have permission
			return false, nil
		}
		return false, fmt.Errorf("error opening path %s: %w", pth, err)
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return false, fmt.Errorf("error getting stat for %s: %w", pth, err)
	}
	return s.IsDir(), nil
}

// listDirEntries wraps a open/close of a path descriptor, thus each readDir
// gorotine can have only one descriptor opened at a time
func listDirEntries(dir string) ([]string, error) {
	d, err := os.Open(dir)
	if err != nil {
		return []string{}, fmt.Errorf("error opening directory %s: %w", dir, err)
	}
	defer d.Close()

	var paths []string
	for {
		ls, err := d.Readdirnames(maxReadDirEntries)
		if err == io.EOF {
			break
		}
		if err != nil {
			return []string{}, fmt.Errorf("error reading diretcory %s: %w", dir, err)
		}
		for _, pth := range ls {
			paths = append(paths, filepath.Join(dir, pth))
		}
	}
	return paths, nil
}

type readDirTask struct {
	dir    string
	queue  chan string
	paths  chan string
	errors chan error
	wg     *sync.WaitGroup
}

func (t *readDirTask) readDir(dir string) {
	defer t.wg.Done()

	ls, err := listDirEntries(dir)
	if err != nil {
		t.errors <- fmt.Errorf("error reading directory %s: %w", dir, err)
		return
	}
	for _, pth := range ls {
		isDir, err := isDir(pth)
		if err != nil {
			t.errors <- fmt.Errorf("error opening %s: %w", pth, err)
			break
		}
		if isDir {
			t.wg.Add(1)
			t.queue <- pth
			continue
		}
		if isValidJSON(pth) {
			t.paths <- pth
		} else {
			log.Output(2, fmt.Sprintf("Invalid JSON file path for a CNPJ: %s", pth))
		}
	}
	return
}

func (t *readDirTask) consumer() {
	for dir := range t.queue {
		t.readDir(dir)
	}
}

func allJSONFiles(dir string, paths chan string, errors chan error) {
	var wg sync.WaitGroup
	t := readDirTask{
		dir:    dir,
		queue:  make(chan string, maxReadDirEntries),
		paths:  paths,
		errors: errors,
		wg:     &wg,
	}
	for i := 0; i <= transform.MaxFilesOpened; i++ {
		go t.consumer()
	}
	wg.Add(1)
	t.queue <- dir
	wg.Wait()
	close(paths)
}
