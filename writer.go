package harlog

import (
	"encoding/json"
	"net/http"
	"os"
)

func (h *HARLogger) saveHAR(req *http.Request, entry *HAREntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(h.outputDir, 0755); err != nil {
		return err
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

	filename := h.fileNameFn(req)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(har)
}
