package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var ErrConnection = errors.New("couldn't connect to Anki")
var modelName_AddAnkiCLI = "ankitango"

var addCmd = &cobra.Command{
	Use:   "add [words] [deckName]",
	Short: "Add word to Anki",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")

		var wordsArray []string
		var deckName string

		if filePath != "" {
			if len(args) < 1 {
				fmt.Println("Error: Please specify the deckname.")
				return
			}
			wordsArray = readWord(filePath)
			deckName = args[0]

		} else {
			if len(args) < 2 {
				fmt.Println("Error: Please specify both word and deckname.")
				return
			}
			wordsArray = []string{args[0]}
			deckName = args[1]
		}

		if !checkAnkiRunning() {
			return
		}
		if !isDeck(getDeckName(), deckName) {
			return
		}
		if !IsModel() {
			addNewModel()
		}

		var failedWords []string
		for i := 0; i < len(wordsArray); i++ {
			word := wordsArray[i]

			exists, err := isNote(deckName, word)
			if err != nil {
				fmt.Println("Connection failed isNote ->", word)
				failedWords = append(failedWords, word)
				continue
			}
			if exists {
				continue
				// return
			}
			fields := generateWord(wordsArray[i])
			if len(fields) == 0 {
				continue
				//return
			}
			if err := addCard(fields, deckName); err != nil {
				if errors.Is(err, ErrConnection) {
					fmt.Println("Connection failed addCard ->", word)
					failedWords = append(failedWords, word)
				} else {
					fmt.Println("Skipped ->", word, ":", err)
				}

				continue
			}
		}

		if len(failedWords) != 0 {
			dir := filepath.Dir(filePath)
			outPath := filepath.Join(dir, "fail.txt")
			fail(failedWords, outPath)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all deck in your Anki",
	Run: func(cmd *cobra.Command, args []string) {
		if !checkAnkiRunning() {
			return
		}

		printList(getDeckName())
	},
}

func ankiInvoke(req any) ([]byte, error) {
	url := "http://127.0.0.1:8765"
	jsonData, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	var resp *http.Response
	for i := 0; i < 3; i++ {
		// send request to localhost
		r, err := http.Post(
			url,                       // where
			"application/json",        // which data type(json)
			bytes.NewBuffer(jsonData), // what data
		)
		if err == nil {
			resp = r
			break
		}

		time.Sleep(500 * time.Millisecond)
	}
	if resp == nil {
		return nil, fmt.Errorf("couldn't connect")
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// add new card to Anki deck
func addCard(fields map[string]string, deckName string) error {

	// 1 making json structure in to out
	type Options struct {
		AllowDuplicate bool   `json:"allowDuplicate"`
		DuplicateScope string `json:"duplicateScope"`
	}
	type Note struct {
		DeckName  string            `json:"deckName"`
		ModelName string            `json:"modelName"`
		Fields    map[string]string `json:"fields"`
		Options   Options           `json:"options"`
	}

	type Params struct {
		Note Note `json:"note"`
	}

	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  Params `json:"params"`
	}

	req := AnkiRequest{
		Action:  "addNote",
		Version: 6,
		Params: Params{
			Note: Note{
				DeckName:  deckName,
				ModelName: modelName_AddAnkiCLI,
				Fields:    fields,
				Options: Options{
					AllowDuplicate: false,
					DuplicateScope: "deck",
				},
			},
		},
	}

	body, err := ankiInvoke(req)
	if err != nil {
		return err
	}

	type AnkiResponse struct {
		Result *int64  `json:"result"`
		Error  *string `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	// say error or success
	if ankiResp.Error != nil {
		return fmt.Errorf("Error: %s", *ankiResp.Error)
	}
	fmt.Println("Success!")
	return nil
}
