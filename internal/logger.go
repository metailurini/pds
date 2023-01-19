package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ParallelStdout is the wrapped struct of *os.File to handle logic show log to the console
type ParallelStdout struct {
	*os.File
}

// NewParallelStdout used to init parallel logger: one for stdout console, two is for file log
func NewParallelStdout(logFileName string) (*ParallelStdout, error) {
	stdoutLog, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "OpenFile")
	}
	return &ParallelStdout{
		File: stdoutLog,
	}, nil
}

// Write will show log to the console and run the parent `Write` implementation
func (ps *ParallelStdout) Write(p []byte) (n int, err error) {
	fmt.Printf("%s", p)
	return ps.File.Write(p)
}

// InitLogger is used to construct the zap logger for logging info
func InitLogger(filename string) (*zap.Logger, error) {
	parallelStdout, err := NewParallelStdout(filename)
	if err != nil {
		return nil, errors.Wrap(err, "NewParallelStdout")
	}

	log.Printf("=== start logging at %s ===\n", filename)

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	writer := zapcore.AddSync(parallelStdout)
	defaultLogLevel := zapcore.DebugLevel
	fileEncoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewTee(zapcore.NewCore(fileEncoder, writer, defaultLogLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
