package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
)

var modelName_AddAnkiCLI = "AddAnkiCLI"

var addCmd = &cobra.Command{
	Use:   "add [word] [deckName]",
	Short: "Add word to Anki",
	Run: func(cmd *cobra.Command, args []string) {
		word := args[0]
		deckName := args[1]

		if !checkAnkiRunning() {
			return
		}
		if !isDeck(getDeckName(), deckName) {
			return
		}
		if !IsModel() {
			addNewModel()
		}
		addCard(generateWard(word, "English", "Japanese"), deckName)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all deck in your Anki",
	Run: func(cmd *cobra.Command, args []string) {
		printList(getDeckName())
	},
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
func generateWard(word string, fromLang string, toLang string) map[string]string {
	for {
		url := "https://api.openai.com/v1/chat/completions"
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		openai_apikey := os.Getenv("OPENAI_API_KEY")

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
	url := "http://127.0.0.1:8765"
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
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
			fmt.Println("aruyo---")
			return true
		}
		// fmt.Println(models)
	}
	fmt.Println("naiyoooo")
	return false
}

// add new model
// before run this function, user needs check "AddAnkiCLI" is not available
func addNewModel() {
	url := "http://127.0.0.1:8765"
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
			InOrderFields: []string{"Front", "Front_Sentence", "Back", "Back_Sentence"},
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
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	fmt.Println("model tukuttayo")

	// fmt.Println(string(body))
}

// add new card to Anki deck
func addCard(fields map[string]string, deckName string) {
	url := "http://127.0.0.1:8765"

	// 1 making json structure in to out
	type Note struct {
		DeckName  string            `json:"deckName"`
		ModelName string            `json:"modelName"`
		Fields    map[string]string `json:"fields"`
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
			},
		},
	}

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
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	type AnkiResponse struct {
		Result *int64  `json:"result"`
		Error  *string `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	// say error or success
	// if ankiResp.Error != nil {
	// 	fmt.Println("Error: ", *ankiResp.Error)
	// 	return
	// }
	fmt.Println("Succsess!")
}

func getDeckName() []string {
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
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	// for _, deck := range ankiResp.Result {
	// 	fmt.Println(deck)
	// }
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
