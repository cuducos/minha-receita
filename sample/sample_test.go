package sample

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSampleTargetHaveAllFiles(t *testing.T) {
    data_files := listFilesFromDir("../data")
    sample_files := listFilesFromDir("../sample_data")
    assertArraysHaveSameItems(t, data_files, sample_files)
}

func listFilesFromDir(dir string) []string {
    var files []string
    path, _ := filepath.Abs(dir)
    paths, err := ioutil.ReadDir(path)
    if err != nil {
        panic(err)
    }

    for _, file := range paths {
        if ! file.IsDir() {
            files = append(files, file.Name())
        }
    }
    return files
}

func assertArraysHaveSameItems(t *testing.T, a, b []string) {
        if len(a) != len(b) {
            t.Errorf("Arrays have different lengths")
        }
        if ! reflect.DeepEqual(a, b) {
            t.Errorf("File not found in sample")
        }
    }
