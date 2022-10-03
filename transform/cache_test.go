package transform

import (
	"path/filepath"
	"testing"
)

func TestPathFor(t *testing.T) {
	tmp := t.TempDir()
	c := cache{dir: tmp}

	t.Run("test cnpj key", func(t *testing.T) {
		expected := filepath.Join(tmp, "12", "345", "678", "9012", "34")
		if got := c.pathFor("12345678901234"); got != expected {
			t.Errorf("expected path to be %s, got %s", expected, got)
		}
	})
	t.Run("test non-cnpj key", func(t *testing.T) {
		expected := filepath.Join(tmp, "42")
		if got := c.pathFor("42"); got != filepath.Join(tmp, "42") {
			t.Errorf("expected path to be %s, got %s", expected, got)
		}
	})
}

func TestCheck(t *testing.T) {
	c := cache{dir: t.TempDir()}
	exist, err := c.check("42")
	if err != nil {
		t.Errorf("expected no error checking a new key, got %s", err)
	}
	if exist {
		t.Errorf("expected a new key not to exist, but it does")
	}
	exist, err = c.check("42")
	if err != nil {
		t.Errorf("expected no error checking a key, got %s", err)
	}
	if !exist {
		t.Errorf("expected a checked key to exist, but it does not")
	}
	if err := c.save("fourty-two"); err != nil {
		t.Errorf("expected no error saving a key, got %s", err)
	}
	exist, err = c.check("fourty-two")
	if err != nil {
		t.Errorf("expected no error checking a key, got %s", err)
	}
	if !exist {
		t.Errorf("expected a saved key to exist, but it does not")
	}
}

func TestSave(t *testing.T) {
	c := cache{dir: t.TempDir()}
	if err := c.save("42"); err != nil {
		t.Errorf("expected no error saving a key, got %s", err)
	}
	exist, err := c.check("42")
	if err != nil {
		t.Errorf("expected no error checking a key, got %s", err)
	}
	if !exist {
		t.Errorf("expected a saved key to exist, but it does not")
	}
}
