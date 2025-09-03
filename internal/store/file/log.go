package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LogStore struct {
	workDir string
}

func NewLogStore(workDir string) *LogStore {
	return &LogStore{
		workDir: workDir,
	}
}

func (s *LogStore) NewLogWriter(operationType string, operationID uint) (io.WriteCloser, error) {
	logsDir := filepath.Join(s.workDir, "logs")
	err := os.MkdirAll(logsDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	filename := fmt.Sprintf("%s-%d.log", operationType, operationID)
	logPath := filepath.Join(s.workDir, "logs", filename)

	file, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return file, nil
}

func (s *LogStore) StoreLog(ctx context.Context, operationType string, operationID uint, r io.Reader) error {
	logsDir := filepath.Join(s.workDir, "logs")
	err := os.MkdirAll(logsDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	filename := fmt.Sprintf("%s-%d.log", operationType, operationID)
	logPath := filepath.Join(s.workDir, "logs", filename)

	file, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	if err != nil {
		return fmt.Errorf("failed to copy log content to file: %w", err)
	}

	return nil
}

func (s *LogStore) LoadLog(ctx context.Context, operationType string, operationID uint) (io.ReadCloser, error) {
	filename := fmt.Sprintf("%s-%d.log", operationType, operationID)
	logPath := filepath.Join(s.workDir, "logs", filename)

	file, err := os.Open(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("log file not found for %s ID %d", operationType, operationID)
		}
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}
