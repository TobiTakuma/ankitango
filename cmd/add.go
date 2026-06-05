package cmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// help function

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

func isNote(deckName string, word string) (bool, error) {
	// 1 make request structure
	type Params struct {
		Query string `json:"query"`
	}
	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  Params `json:"params"`
	}
	// 2 making data

	query := fmt.Sprintf(`deck:"%s" Front:"%s"`, deckName, word)
	req := AnkiRequest{
		Action:  "findNotes",
		Version: 6,
		Params: Params{
			Query: query,
		},
	}

	body, err := ankiInvoke(req)
	if err != nil {
		return true, err
		// panic(err)
	}

	type AnkiResponse struct {
		Result []int64 `json:"result"`
		Error  *string `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	if len(ankiResp.Result) > 0 {
		fmt.Println("Error: already exists →", word)
		return true, nil
	}
	return false, nil
}

func printList(list []string) {
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i])
	}
}

func isDeck(deckList []string, deckName string) bool {
	for i := 0; i < len(deckList); i++ {
		if deckList[i] == deckName {
			return true
		}
	}
	fmt.Println("Error:  deck was not found: " + deckName)
	fmt.Println(`If the deck name contains spaces, enclose it in double quotes.
				Example: ankitango add “hello world” “deckName”`)
	return false
}

// generate new words
func generateWord(word string) map[string]string {
	cfg := loadConfig() // import config setting
	// if apikey has not been configured, return error
	if cfg.APIKey == "" {
		fmt.Println("Error: No API key has been configured. \nPlease configure it using `ankitango config apikey <key>`.")
		return map[string]string{}
	}
	if (cfg.FromLang == "") || (cfg.ToLang == "") {
		fmt.Println("Error: No language has been configured. \nPlease configure it using ankitango config apikey <from> <to>`.\nExample) if you want to from english to japanese. \n`ankitango config lang English Japanese`")
		return map[string]string{}
	}
	openai_apikey := cfg.APIKey
	fromLang := cfg.FromLang
	toLang := cfg.ToLang
	fmt.Println("\nGenerating...")

	for {
		url := "https://api.openai.com/v1/chat/completions"

		// json structure for request
		type Messages struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		type Response_format struct {
			Type string `json:"type"`
		}
		type OpenAIRequest struct {
			Model           string          `json:"model"`
			Messages        []Messages      `json:"messages"`
			Response_format Response_format `json:"response_format"`
		}

		// json structure for responce
		type OpenAIMessage struct {
			Content string `json:"content"`
		}
		type Choice struct {
			Message OpenAIMessage `json:"message"`
		}
		type OpenAIResponse struct {
			Choices []Choice `json:"choices"`
		}

		content := fmt.Sprintf(
			"Given a word, return ONLY a JSON object with exactly these 4 keys:\n"+
				"- \"Front\": the word in %s\n"+
				"- \"Front_Sentence\": a natural example sentence in %s using the word\n"+
				"- \"Back\": the translation of the word in %s\n"+
				"- \"Back_Sentence\": the translation of the example sentence in %s\n"+
				"No markdown, no explanation, only the JSON object.",
			fromLang, fromLang, toLang, toLang,
		)

		opai_req := OpenAIRequest{
			Model: "gpt-4o-mini",
			Messages: []Messages{
				{Role: "system", Content: content},
				{Role: "user", Content: word},
			},
			Response_format: Response_format{Type: "json_object"},
		}
		jsonData, _ := json.Marshal(opai_req)

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+openai_apikey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var openAIResp OpenAIResponse
		json.Unmarshal(body, &openAIResp)

		if len(openAIResp.Choices) == 0 {
			fmt.Println("Error: The responce from OpenAI is incorrect.\nCheck your API key or network")
			return map[string]string{}
		}

		// contentを取り出す
		result := openAIResp.Choices[0].Message.Content

		fmt.Println(result)
		var fields map[string]string
		json.Unmarshal([]byte(result), &fields)
		if fields != nil {
			return fields
		}
	}
}

// check if the model exists
func IsModel() bool {
	// 1 make request structure
	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
	}
	// 2 making data
	req := AnkiRequest{
		Action:  "modelNames",
		Version: 6,
	}

	body, err := ankiInvoke(req)
	if err != nil {
		panic(err)
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	for _, models := range ankiResp.Result {
		if string(models) == modelName_AddAnkiCLI {
			return true
		}
	}
	fmt.Println("The specified model does not exist.\nCreating a new one...")
	return false
}

// add new model
// before run this function, user needs check "AddAnkiCLI" is not available
func addNewModel() {
	// 1 make json structure
	type CardTemplates struct {
		Name  string `json:"Name"`
		Front string `json:"Front"`
		Back  string `json:"Back"`
	}
	type Params struct {
		ModelName     string          `json:"modelName"`
		InOrderFields []string        `json:"inOrderFields"`
		IsClose       bool            `json:"isCloze"`
		CSS           string          `json:"css"`
		CardTemplates []CardTemplates `json:"cardTemplates"`
	}
	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  Params `json:"params"`
	}

	// 2 make request
	req := AnkiRequest{
		Action:  "createModel",
		Version: 6,
		Params: Params{
			ModelName:     modelName_AddAnkiCLI,
			InOrderFields: []string{"Front", "Front_Sentence", "Back", "Back_Sentence", "Pronunciation", "Audio", "Synonym", "Note"},
			IsClose:       false,
			CSS:           "",
			CardTemplates: []CardTemplates{
				{
					Name:  "Card1",
					Front: `{{Front}}<p><div style='font-family: "Arial"; font-size: 20px;'>{{Front_Sentence}}</div>`,
					Back:  `{{FrontSide}}<hr id=answer>{{Back}}<p><div style='font-family: "Arial"; font-size: 20px;'>{{Back_Sentence}}</div>`,
				},
				{
					Name:  "Card2",
					Front: `{{Back}}<p><div style='font-family: "Arial"; font-size: 20px;'>{{Back_Sentence}}</div>`,
					Back:  `{{FrontSide}}<hr id=answer>{{Front}}<p><div style='font-family: "Arial"; font-size: 20px;'>{{Front_Sentence}}</div>`,
				},
			},
		},
	}

	body, err := ankiInvoke(req)
	if err != nil {
		panic(err)
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	fmt.Println("A new model named “ankitango” has been created.")
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

func getDeckName() []string {
	// 1 make request structure
	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
	}
	// 2 making data
	req := AnkiRequest{
		Action:  "deckNames",
		Version: 6,
	}

	body, err := ankiInvoke(req)
	if err != nil {
		panic(err)
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	return ankiResp.Result
}

// check if
func checkAnkiRunning() bool {
	url := "http://127.0.0.1:8765"
	// 1 make request structure
	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
	}
	// 2 making data
	req := AnkiRequest{
		Action:  "deckNames",
		Version: 6,
	}

	// 3 convert to JSON. Marshal(変換)
	jsonData, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	// 4 send request to localhost
	resp, err := http.Post(
		url,                       // where
		"application/json",        // which data type(json)
		bytes.NewBuffer(jsonData), // what data
	)
	// if Anki isn't run it return error
	if err != nil {
		fmt.Println("Error: Anki is not ruuning")
		return false
	}

	// if get response, Anki is already running
	defer resp.Body.Close()
	// fmt.Println("Anki is running")
	return true
}
