package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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
