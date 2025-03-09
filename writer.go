package harlog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (l *Logger) saveHAR(req *http.Request, entry *HAREntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(l.outputDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get filename and validate it's within the output directory
	filename := l.fileNameFn(req)
	absOutputDir, err := filepath.Abs(l.outputDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of output directory: %w", err)
	}
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of file: %w", err)
	}

	// Clean paths and ensure they are in canonical form
	absOutputDir = filepath.Clean(absOutputDir)
	absFilename = filepath.Clean(absFilename)

	// Check if the file path is within the output directory
	rel, err := filepath.Rel(absOutputDir, absFilename)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return fmt.Errorf("file path %s is outside of output directory %s", filename, l.outputDir)
	}

	har := HAR{
		Log: HARLog{
			Version: "1.2",
			Creator: HARCreator{
				Name:    "harlog",
				Version: "1.0",
			},
			Entries: []HAREntry{*entry},
		},
	}

	file, err := os.OpenFile(absFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(har); err != nil {
		return fmt.Errorf("failed to encode HAR: %w", err)
	}

	return nil
}
