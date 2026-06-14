// Package logging is designed to be used modular with the server. It should serve debugging and productional uses
// filesizehandling like "lumberjack" is not yet implemented
package logging

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

// GlobalLogger acts as the fallback instance.
var GlobalLogger *slog.Logger

// GSRotatingLogWriter handles size-based log rotation using only the standard library.
type GSRotatingLogWriter struct {
	Filename string
	MaxSize  int64 // In bytes
	mu       sync.Mutex
	file     *os.File
	currSize int64
}

// LogSession holds the file handle and the logger together to ensure
// that the file can be closed gracefully by the main function.
type LogSession struct {
	File   *io.Closer
	Writer *GSRotatingLogWriter
	Logger *slog.Logger
}

// Write checks the file size before performing the actual write operation.
func (rlw *GSRotatingLogWriter) Write(p []byte) (n int, err error) {
	rlw.mu.Lock()
	defer rlw.mu.Unlock()

	writeLen := int64(len(p))

	// If writing this would exceed MaxSize, rotate the file.
	if rlw.file != nil && rlw.currSize+writeLen > rlw.MaxSize {
		if err := rlw.rotate(); err != nil {
			// Fallback: if rotation fails, we still try to write to the current file
			// to avoid losing log data.
			os.Stderr.WriteString("Log rotation failed: " + err.Error() + "\n")
		}
	}

	n, err = rlw.file.Write(p)
	rlw.currSize += int64(n)
	return n, err
}

// rotate closes the current file, renames it, and opens a new one.
func (rlw *GSRotatingLogWriter) rotate() error {
	if rlw.file != nil {
		rlw.file.Close()
	}

	// Simple rotation: app.log -> app.log.1
	// For now, we only keep one backup to stay simple, but this can be expanded.
	backupName := rlw.Filename + ".1"
	_ = os.Remove(backupName) // Remove old backup if it exists
	if err := os.Rename(rlw.Filename, backupName); err != nil {
		return err
	}

	// Open a fresh file
	f, err := os.OpenFile(rlw.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	rlw.file = f
	rlw.currSize = 0
	return nil
}

// Close ensures the file is shut down correctly.
func (rlw *GSRotatingLogWriter) Close() error {
	rlw.mu.Lock()
	defer rlw.mu.Unlock()
	if rlw.file != nil {
		return rlw.file.Close()
	}
	return nil
}

// SetupLogger initializes a structured JSON logger.
// It returns a LogSession so the caller can 'defer' the closing of the log file.
func SetupLogger(fileName string, level slog.Level, maxSize int64) (*LogSession, error) {
	// 1. Initialize the rotating writer
	rotator := &GSRotatingLogWriter{
		Filename: fileName,
		MaxSize:  maxSize,
	}

	// Open the initial file or create it
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	info, _ := f.Stat()
	rotator.file = f
	rotator.currSize = info.Size()

	// 2. MultiWriter sends logs to both terminal and the rotating file.
	multiWriter := io.MultiWriter(os.Stdout, rotator)

	handlerOptions := &slog.HandlerOptions{Level: level}
	jsonHandler := slog.NewJSONHandler(multiWriter, handlerOptions)
	newLogger := slog.New(jsonHandler)

	GlobalLogger = newLogger
	slog.SetDefault(newLogger)

	return &LogSession{
		Writer: rotator, // Updated to hold our rotator for closing
		Logger: newLogger,
	}, nil
}

// Get returns a logger with attribute.
// This makes filtering logs in large projects significantly easier.
func Get(componentName string) *slog.Logger {
	if GlobalLogger == nil {
		// Fallback to standard default if SetupLogger hasn't been called.
		return slog.Default().With("component", componentName)
	}
	return GlobalLogger.With("component", componentName)
}
