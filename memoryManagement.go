package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

type MountedFile struct {
	Path     string   `json:"path"`
	Cols     []string `json:"cols"`
	RowCount int      `json:"rowCount"`
}

var columnMap = make(map[string][]string)
var contentMap = make(map[string][][]string)

func mountFile(path string) error {
	if filepath.Ext(path) != ".csv" {
		return fmt.Errorf("not a .csv file: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	columnMap[path] = records[0]
	contentMap[path] = records[1:]

	return nil
}

func unmountFile(path string) {
	delete(columnMap, path)
	delete(contentMap, path)
}

func getStatus() ([]MountedFile, error) {
	var mountedFiles []MountedFile
	for path, cols := range columnMap {
		content, ok := contentMap[path]
		if !ok {
			return nil, fmt.Errorf("no content for path: %s", path)
		}
		mountedFiles = append(mountedFiles, MountedFile{
			Path:     path,
			Cols:     cols,
			RowCount: len(content),
		})
	}
	return mountedFiles, nil
}
