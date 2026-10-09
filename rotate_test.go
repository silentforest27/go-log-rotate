package logrotate

import (
	"os"
	"strings"
	"testing"
)

func TestRotation(t *testing.T) {
	filename := "test.log"
	defer os.Remove(filename)
	defer os.RemoveAll("test.log.*")

	// Max size 10 bytes, max backups 2
	r, err := NewRotator(filename, 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// Trigger multiple rotations
	for i := 0; i < 5; i++ {
		_, err = r.Write([]byte("this is a long string"))
		if err != nil {
			t.Fatal(err)
		}
	}

	files, _ := os.ReadDir(".")
	backupCount := 0
	for _, f := range files {
		if f.Name() != filename && strings.HasPrefix(f.Name(), "test.log.") {
			backupCount++
		}
	}

	if backupCount > 2 {
		t.Errorf("Expected at most 2 backup files, found %d", backupCount)
	}
	if backupCount == 0 {
		t.Error("Expected rotated log files to be created")
	}

	// Verify at least one file is gzipped
	foundGzip := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gz") {
			foundGzip = true
			break
		}
	}
	if !foundGzip {
		t.Error("Expected rotated log files to be compressed with gzip")
	}
}