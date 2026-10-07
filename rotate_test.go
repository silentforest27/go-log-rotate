package logrotate

import (
	"os"
	"testing"
)

func TestRotation(t *testing.T) {
	filename := "test.log"
	defer os.Remove(filename)
	defer os.RemoveAll("test.log.*")

	// Max size 10 bytes to trigger rotation quickly
	r, err := NewRotator(filename, 10)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// First write: fits in limit
	_, err = r.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}

	// Second write: triggers rotation
	_, err = r.Write([]byte(" world this is long"))
	if err != nil {
		t.Fatal(err)
	}

	files, _ := os.ReadDir(".")
	foundBackup := false
	for _, f := range files {
		if f.Name() != filename && len(f.Name()) > 8 && f.Name()[:8] == "test.log" {
			foundBackup = true
			break
		}
	}

	if !foundBackup {
		t.Error("Expected rotated log file to be created")
	}
}