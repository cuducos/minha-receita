package download

import "testing"

func TestGetFiles(t *testing.T) {
	fs := getFiles(false, "data")
	if len(fs) != files+1 {
		t.Errorf("Expected %d files, got %d", files+1, len(fs))
	}
}
