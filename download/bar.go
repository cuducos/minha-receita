package download

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

// wrapper around progressbar.ProgressBar to allow a bytes progress
// downloadProgressBar with a descriptions that shows how many files have been
// donwloaded and how many are still pending.
type downloadProgressBar struct {
	main        *progressbar.ProgressBar
	total       int
	done        int
	updateBytes chan int64
	updateTotal chan struct{}
}

func (b *downloadProgressBar) Write(d []byte) (int, error) {
	b.updateBytes <- int64(len(d))
	return len(d), nil
}

func (b *downloadProgressBar) description() string {
	return fmt.Sprintf("Downloading (%d of %d files done)", b.done, b.total)
}

func (b *downloadProgressBar) update() {
	b.main.Describe(b.description())
}

func (b *downloadProgressBar) addBytes(n int64) {
	// on the last file, the file count must be updated before adding bytes,
	// otherwise the description won't update
	if b.done == b.total-1 {
		if b.main.State().CurrentBytes+float64(n) == float64(b.main.GetMax64()) {
			b.done = b.total
			b.update()
		}
	}
	b.main.Add64(n)
}

func (b *downloadProgressBar) addFile() {
	b.done++
	b.update()
}
