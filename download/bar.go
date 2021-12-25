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

func (b *downloadProgressBar) isFinished() bool {
	return b.total == b.done
}

func (b *downloadProgressBar) Write(d []byte) (int, error) {
	b.updateBytes <- int64(len(d))
	return len(d), nil
}

func (b *downloadProgressBar) description() string {
	return fmt.Sprintf("Downloading (%d of %d files done)", b.done, b.total)
}
