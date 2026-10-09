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

func TestManualRotation(t *testing.T) {
	filename := "manual_test.log"
	defer os.Remove(filename)
	defer os.RemoveAll("manual_test.log.*")

	r, err := NewRotator(filename, 1024, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	_, _ = r.Write([]byte("some data"))

	if err := r.Rotate(); err != nil {
		t.Fatalf("Manual rotate failed: %v", err)
	}

	files, _ := os.ReadDir(".")
	foundBackup := false
	for _, f := range files {
		if f.Name() != filename && strings.HasPrefix(f.Name(), "manual_test.log.") {
			foundBackup = true
			break
		}
	}

	if !foundBackup {
		t.Error("Expected a backup file after manual rotation")
	}
}

func TestDisableCompression(t *testing.T) {
	filename := "no_comp.log"
	defer os.Remove(filename)
	defer os.RemoveAll("no_comp.log.*")

	r, err := NewRotator(filename, 10, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	r.SetCompress(false)
	_, _ = r.Write([]byte("this string is long"))

	files, _ := os.ReadDir(".")
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gz") && strings.HasPrefix(f.Name(), "no_comp.log") {
			t.Errorf("Expected no gzipped files when compression is disabled, found %s", f.Name())
		}
	}
}

func TestSizeMethod(t *testing.T) {
	filename := "size_test.log"
	defer os.Remove(filename)
	defer os.RemoveAll("size_test.log.*")

	r, err := NewRotator(filename, 1024, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data := []byte("hello")
	_, _ = r.Write(data)

	if r.Size() != int64(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), r.Size())
	}
}