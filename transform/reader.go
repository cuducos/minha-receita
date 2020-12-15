package transform

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"

	"github.com/cuducos/minha-receita/download"
)

type zippedFile struct {
	path      string
	archive   *zip.ReadCloser
	firstFile io.ReadCloser
}

func (z *zippedFile) Close() error {
	if z.firstFile != nil {
		err := z.firstFile.Close()
		if err != nil {
			return err
		}
	}
	if z.archive != nil {
		err := z.archive.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func newZippedFile(dir string, i int) (*zippedFile, error) {
	p := filepath.Join(dir, fmt.Sprintf(download.FilePattern, i))
	z, err := zip.OpenReader(p)
	if err != nil {
		return nil, err
	}

	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			z.Close()
			return nil, err
		}
		return &zippedFile{p, z, r}, nil
	}
	z.Close()
	return nil, fmt.Errorf("no zipped file found in %s", p)
}
