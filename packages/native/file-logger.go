package native

import (
	"os"
	"sync"
)

// FileLogger writes log messages to a file
type FileLogger struct {
	mu   sync.Mutex
	file *os.File
}

// NewFileLogger creates a new FileLogger
func NewFileLogger(path string) (*FileLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLogger{file: f}, nil
}

// Write writes a log message
func (l *FileLogger) Write(message string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file == nil {
		return nil
	}
	_, err := l.file.WriteString(message + "\n")
	return err
}

// Close closes the log file
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		err := l.file.Close()
		l.file = nil
		return err
	}
	return nil
}
