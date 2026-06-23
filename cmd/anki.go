package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// get all notes from deck
func findNotes(deckName string) ([]int, error) {
	type Params struct {
		Query string `json:"query"`
	}

	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  Params `json:"params"`
	}

	query := "deck:" + deckName
	req := AnkiRequest{
		Action:  "findNotes",
		Version: 6,
		Params: Params{
			Query: query,
		},
	}

	body, err := ankiInvoke(req)
	if err != nil {
		return nil, err
	}

	type AnkiResponse struct {
		Result []int   `json:"result"`
		Error  *string `json:"error"`
	}

	var ankiResp AnkiResponse
	json.Unmarshal(body, &ankiResp)
	return ankiResp.Result, nil
}

func noteInfo(notesId int) (string, error) {
	type Params struct {
		Notes []int `json:"notes"`
	}

	type AnkiRequest struct {
		Action  string `json:"action"`
		Version int    `json:"version"`
		Params  Params `json:"params"`
	}

	ids := []int{notesId}

	req := AnkiRequest{
		Action:  "notesInfo",
		Version: 6,
		Params: Params{
			Notes: ids,
		},
	}
	body, err := ankiInvoke(req)
	if err != nil {
		return "", err
	}

	type FieldInfo struct {
		Value string `json:"value"`
		Order int    `json:"order"`
	}

	type NoteInfo struct {
		NoteID int                  `json:"noteId"`
		Fields map[string]FieldInfo `json:"fields"`
	}

	type AnkiResponse struct {
		Result []NoteInfo `json:"result"`
		Error  *string    `json:"error"`
	}

	var ankiResp AnkiResponse
	json.Unmarshal(body, &ankiResp)

	if len(ankiResp.Result) == 0 {
		return "", nil
	}

	front := ankiResp.Result[0].Fields["Front"].Value

	fmt.Println("\nFront         :", ankiResp.Result[0].Fields["Front"].Value)
	fmt.Println("Back          :", ankiResp.Result[0].Fields["Front_Sentence"].Value)
	fmt.Println("Front Sentence:", ankiResp.Result[0].Fields["Back"].Value)
	fmt.Println("Back Sentence :", ankiResp.Result[0].Fields["Back_Sentence"].Value)

	return front, nil
}

func updateNoteFields(noteId int, fields map[string]string) error {
	type Note struct {
		ID     int               `json:"id"`
		Fields map[string]string `json:"fields"`
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
		Action:  "updateNoteFields",
		Version: 6,
		Params: Params{
			Note: Note{
				ID:     noteId,
				Fields: fields,
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

// send request to anki(localhost)
// try 3 times when an ankiconnect error occurs
func ankiInvoke(req any) ([]byte, error) {
	url := "http://127.0.0.1:8765"
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
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

func isDeck(deckList []string, deckName string) error {
	for i := 0; i < len(deckList); i++ {
		if deckList[i] == deckName {
			return nil
		}
	}
	return fmt.Errorf("deck was not found: %s\nIf the deck name contains spaces, enclose it in double quotes.\nExample: ankitango add \"hello world\" \"deckName\"", deckName)
}

// check if the model exists
func IsModel() (bool, error) {
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
		return false, err
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	for _, models := range ankiResp.Result {
		if string(models) == modelName_AddAnkiCLI {
			return true, nil
		}
	}
	fmt.Println("The specified model does not exist.\nCreating a new one...")
	return false, nil
}

// add new model
// before run this function, user needs check "AddAnkiCLI" is not available
func addNewModel() error {
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
		return err
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	fmt.Println("A new model named “ankitango” has been created.")
	return nil
}

func getDeckName() ([]string, error) {
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
		return nil, err
	}

	type AnkiResponse struct {
		Result []string `json:"result"`
		Error  *string  `json:"error"`
	}

	var ankiResp AnkiResponse       // 空の入れ物を用意
	json.Unmarshal(body, &ankiResp) // JSONをGoのデータに流し込む

	return ankiResp.Result, nil
}

// check if
func checkAnkiRunning() error {
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
		return fmt.Errorf("failed to build request: %w", err)
	}

	// 4 send request to localhost
	resp, err := http.Post(
		url,                       // where
		"application/json",        // which data type(json)
		bytes.NewBuffer(jsonData), // what data
	)
	// if Anki isn't run it return error
	if err != nil {
		return fmt.Errorf("Anki is not running")
	}

	// if get response, Anki is already running
	defer resp.Body.Close()
	// fmt.Println("Anki is running")
	return nil
}
