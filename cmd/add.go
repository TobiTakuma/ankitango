package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var ErrConnection = errors.New("couldn't connect to Anki")
var modelName_AddAnkiCLI = "ankitango"

var addCmd = &cobra.Command{
	Use:   "add [words] [deckName]",
	Short: "Add a word to Anki",
	Run: func(cmd *cobra.Command, args []string) {
		filePath, _ := cmd.Flags().GetString("file")

		var wordsArray []string
		var deckName string

		if filePath != "" {
			if len(args) < 1 {
				fmt.Println("Error: Please specify the deckname.")
				return
			}
			words, err := readWord(filePath)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			wordsArray = words
			deckName = args[0]

		} else {
			if len(args) < 2 {
				fmt.Println("Error: Please specify both word and deckname.")
				fields, _ := generateWord(args[0])
				printLLMresult(fields)
				return
			}
			wordsArray = []string{args[0]}
			deckName = args[1]
		}

		if err := checkAnkiRunning(); err != nil {
			fmt.Println("Error:", err)
			return
		}
		decks, err := getDeckName()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if err := isDeck(decks, deckName); err != nil {
			fmt.Println("Error:", err)
			return
		}
		modelExists, err := IsModel()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if !modelExists {
			if err := addNewModel(); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}

		var failedWords []string
		for i := 0; i < len(wordsArray); i++ {
			if len(wordsArray) > 1 {
				fmt.Println("--------------------")

			}
			fmt.Println("Generating...")
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
			fields, err := generateWord(wordsArray[i])
			if err != nil {
				fmt.Println("Error:", err)
				failedWords = append(failedWords, word)
				continue
			}
			printLLMresult(fields)
			if err := addCard(fields, deckName); err != nil {
				if errors.Is(err, ErrConnection) {
					fmt.Println("Connection failed addCard ->", word)
					failedWords = append(failedWords, word)
				} else {
					fmt.Println("Skipped ->", word, ":", err)
				}

				continue
			}

			fmt.Println("Success!")
			fmt.Printf("[%s] has been added to [%s]\n", word, deckName)
		}

		if len(failedWords) != 0 {
			dir := filepath.Dir(filePath)
			outPath := filepath.Join(dir, "fail.txt")
			if err := fail(failedWords, outPath); err != nil {
				fmt.Println("Error:", err)
			}
		}
	},
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
	return nil
}
