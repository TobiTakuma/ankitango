package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readWord(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error: could not open file: ", path)
		return []string{}
	}
	defer file.Close()

	var words []string

	if filepath.Ext(path) == ".csv" {
		// CSV
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1
		records, err := reader.ReadAll() //[][]string
		if err != nil {
			fmt.Println("Error: failed to read CSV: ", err)
			return []string{}
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
		fmt.Println("Error: not support this file path")
		return []string{}
	}

	return words

}

func fail(failedWords []string, path string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error: could not create fail.txt: ", err)
	}
	defer file.Close()

	for _, w := range failedWords {
		fmt.Fprintln(file, w)
	}

	fmt.Println("failed words were saved to ", path)
}
