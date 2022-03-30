package sample

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSample(t *testing.T) {
    Sample("../data", "../data/sample", 100)
    dataFiles := listFilesFromDir(t, "../data")
    sampleFiles := listFilesFromDir(t, "../data/sample")
    if ! reflect.DeepEqual(dataFiles, sampleFiles) {
        t.Errorf("File not found in sample")
    }
}

func listFilesFromDir(t *testing.T, dir string) []string {
    var files []string
    path, _ := filepath.Abs(dir)
    paths, err := ioutil.ReadDir(path)
    if err != nil {
        t.Errorf("Could not list dir: %s, %v", dir, err)
    }

    for _, file := range paths {
        if ! file.IsDir() {
            files = append(files, file.Name())
        }
    }
    return files
}
