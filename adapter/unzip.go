package adapter

import (
	"archive/zip"
)

func (a *Adapter) unzip(e chan<- error, l chan<- []string, s string) {
	z, err := zip.OpenReader(s)
	if err != nil {
		e <- err
		return
	}
	defer z.Close()

	for _, f := range z.File {
		if err := a.lineProducer(l, f); err != nil {
			e <- err
			return
		}
	}

	e <- nil
}
