package sample

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSample(t *testing.T) {
    src, dst := "../testdata", t.TempDir()
    if err:= Sample(src, dst, 100); err != nil {
        t.Fatalf("expected no error running sample, got %s", err)
    }
    dataFiles := listFilesFromDir(t, src)
    sampleFiles := listFilesFromDir(t, dst)
    if !reflect.DeepEqual(dataFiles, sampleFiles) {
        t.Errorf("File not found in sample")
    }
}

func listFilesFromDir(t *testing.T, dir string) []string {
    ext := []string{"json", "csv", "zip"}
    var files []string
    path, err := filepath.Abs(dir)
    if err != nil {
        t.Errorf("could not read directory %s", dir)
    }
    paths, err := ioutil.ReadDir(path)
    if err != nil {
        t.Errorf("Could not list dir: %s, %v", dir, err)
    }

    for _, file := range paths {
        if !file.IsDir() && contains(ext, strings.ToLower(filepath.Ext(file.Name()))) {
            files = append(files, file.Name())
        }
    }
    return files
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
