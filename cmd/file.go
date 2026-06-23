package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readWord(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %s", path)
	}
	defer file.Close()

	var words []string

	if filepath.Ext(path) == ".csv" {
		// CSV
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll() //[][]string
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}
		for _, record := range records {
			for _, cell := range record {
				w := strings.TrimSpace(cell)
				if w != "" {
					words = append(words, w)
				}
			}
		}

	} else if filepath.Ext(path) == ".txt" {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				words = append(words, line)
			}
		}

	} else {
		return nil, fmt.Errorf("not support this file path: %s", path)
	}

	return words, nil

}

func fail(failedWords []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create fail.txt: %w", err)
	}
	defer file.Close()

	for _, w := range failedWords {
		fmt.Fprintln(file, w)
	}

	fmt.Println("failed words were saved to ", path)
	return nil
}
