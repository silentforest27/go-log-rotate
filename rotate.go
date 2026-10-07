package logrotate

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Rotator handles log rotation based on file size.
type Rotator struct {
	filename    string
	maxSize     int64
	currentSize int64
	file        *os.File
	mu          sync.Mutex
}

// NewRotator initializes a new Rotator instance.
func NewRotator(filename string, maxSize int64) (*Rotator, error) {
	r := &Rotator{
		filename: filename,
		maxSize:  maxSize,
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
	
	// Simple rotation: move current to .1
	backupName := fmt.Sprintf("%s.%d", r.filename, time.Now().Unix())
	if err := os.Rename(r.filename, backupName); err != nil {
		return err
	}
	
	return r.open()
}

// Close closes the underlying log file.
func (r *Rotator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.file.Close()
}