package sample

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSample(t *testing.T) {
    src, dst := "../testdata", "../testdata/sample"
    Sample(src, dst, 100)
    dataFiles := listFilesFromDir(t, src)
    sampleFiles := listFilesFromDir(t, dst)
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
        fileExt := strings.ToLower(filepath.Ext(file.Name()))
        ext := []string{"json", "csv", "zip"}
        if ! file.IsDir() && contains(ext, fileExt) {
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
