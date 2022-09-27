package download

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

type bytesProgress struct {
	path string
	size uint64
}

// wrapper around progressbar.ProgressBar to allow a bytes progress
// downloadProgressBar with a descriptions that shows how many files have been
// donwloaded and how many are still pending.
type downloadProgressBar struct {
	main        *progressbar.ProgressBar
	totalFiles  uint
	filesDone   uint
	downloaded  map[string]uint64
	updateFiles chan struct{}
	updateBytes chan bytesProgress
}

func (b *downloadProgressBar) description() string {
	return fmt.Sprintf("Downloading (%d of %d files done)", b.filesDone, b.totalFiles)
}

func (b *downloadProgressBar) downloadedBytes() uint64 {
	var t uint64
	for _, s := range b.downloaded {
		t += s
	}
	return t
}

func (b *downloadProgressBar) done() bool {
	if b.filesDone != b.totalFiles {
		return false
	}
	if !b.main.IsFinished() {
		b.main.ChangeMax64(int64(b.downloadedBytes()))
		b.main.Finish()
	}
	return true
}

func (b *downloadProgressBar) run() struct{} {
	for {
		if b.done() {
			return struct{}{}
		}
		select {
		case <-b.updateFiles:
			b.filesDone++
			b.main.Describe(b.description())
		case p := <-b.updateBytes:
			b.downloaded[p.path] = p.size
			b.main.Set64(int64(b.downloadedBytes()))
		}
	}
}

func newBar(totalFiles uint, totalBytes uint64, silent bool) *downloadProgressBar {
	bar := downloadProgressBar{
		totalFiles:  totalFiles,
		downloaded:  make(map[string]uint64, totalFiles),
		updateFiles: make(chan struct{}),
		updateBytes: make(chan bytesProgress),
	}
	createBar := progressbar.DefaultBytes
	if silent {
		createBar = progressbar.DefaultBytesSilent
	}
	bar.main = createBar(int64(totalBytes), bar.description())
	return &bar
}
