package logrotate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Rotator handles log rotation based on file size.
type Rotator struct {
	filename    string
	maxSize     int64
	maxBackups  int
	currentSize int64
	file        *os.File
	mu          sync.Mutex
}

// NewRotator initializes a new Rotator instance.
func NewRotator(filename string, maxSize int64, maxBackups int) (*Rotator, error) {
	r := &Rotator{
		filename:   filename,
		maxSize:    maxSize,
		maxBackups:  maxBackups,
	}
	if err := r.open(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Rotator) open() error {
	file, err := os.OpenFile(r.filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		return err
	}
	r.file = file
	r.currentSize = info.Size()
	return nil
}

// Write implements io.Writer
func (r *Rotator) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.currentSize+int64(len(p)) > r.maxSize {
		if err := r.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = r.file.Write(p)
	r.currentSize += int64(n)
	return n, err
}

func (r *Rotator) rotate() error {
	r.file.Close()
	
	backupName := fmt.Sprintf("%s.%d", r.filename, time.Now().Unix())
	if err := os.Rename(r.filename, backupName); err != nil {
		return err
	}
	
	if err := r.prune(); err != nil {
		// Pruning failure shouldn't stop logging, but we log it or return it
		// For this library, we'll return it to let the user decide
		return err
	}
	
	return r.open()
}

func (r *Rotator) prune() error {
	if r.maxBackups <= 0 {
		return nil
	}

	files, err := os.ReadDir(filepath.Dir(r.filename))
	if err != nil {
		return err
	}

	var backups []os.FileInfo
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if strings.HasPrefix(f.Name(), filepath.Base(r.filename)+".") {
			info, err := f.Info()
			if err != nil {
				continue
			}
			backups = append(backups, info)
		}
	}

	if len(backups) <= r.maxBackups {
		return nil
	}

	// Sort by modification time ascending (oldest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime().Before(backups[j].ModTime())
	})

	// Remove oldest files until we hit maxBackups
	toRemove := len(backups) - r.maxBackups
	for i := 0; i < toRemove; i++ {
		path := filepath.Join(filepath.Dir(r.filename), backups[i].Name())
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying log file.
func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}