package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// generateWord に渡すプロンプトのテンプレート。
// %s は順に fromLang, toLang, toLang, fromLang, toLang が入る。
// raw string なので行頭インデント禁止（タブがそのまま本文に入るため左寄せで書く）。
const promptTemplate = `Given a word, return ONLY a JSON object with exactly these 4 keys:

- "Front": the word in %s

- "Back": the meaning(s) of the word in %s.
  Default to ONE meaning. Prefer the fewest senses possible.
  Split with "；" ONLY when the senses are genuinely UNRELATED
  (a different part of speech, e.g. patient=患者/忍耐強い, or an unrelated concept, e.g. bank=銀行/土手).
  Treat different situations, registers, or politeness levels of the SAME idea as ONE sense, NOT a new sense
  (e.g. "hello" is just 挨拶: do NOT split こんにちは and もしもし).
  Use "、" only for synonyms within the same sense, and avoid padding with near-synonyms.
  If multiple senses translate to the SAME word in %s, list that word only ONCE (never repeat it).

- "Front_Sentence": a natural example sentence in %s using the word in its MOST COMMON sense.
  Keep it 6–12 words. Avoid overly simple drills and overly long multi-clause sentences.

- "Back_Sentence": the translation of that sentence in %s, using the SAME sense as Front_Sentence

No markdown, no explanation, only the JSON object.

Example of a word with multiple senses, "patient":
{"Front":"patient","Back":"患者；忍耐強い","Front_Sentence":"The doctor examined the patient carefully.","Back_Sentence":"医者は患者を注意深く診察した。"}

Example of a single-sense word, "apple":
{"Front":"apple","Back":"りんご","Front_Sentence":"He bought fresh apples at the market.","Back_Sentence":"彼は市場で新鮮なリンゴを買った。"}

Example of a word that looks multi-sense but is ONE sense, "hello" (greeting only, do NOT split by situation):
{"Front":"hello","Back":"こんにちは","Front_Sentence":"She said hello to everyone she met.","Back_Sentence":"彼女は会う人みんなに挨拶した。"}`

// generate new words
func generateWord(word string) map[string]string {
	cfg := loadConfig() // import config setting
	// if apikey has not been configured, return error
	if cfg.APIKey == "" {
		fmt.Println("Error: No API key has been configured. \nPlease configure it using `ankitango config apikey <provider> <key>`.")
		return map[string]string{}
	}
	if (cfg.FromLang == "") || (cfg.ToLang == "") {
		fmt.Println("Error: No language has been configured. \nPlease configure it using ankitango config apikey <from> <to>`.\nExample) if you want to from english to japanese. \n`ankitango config lang English Japanese`")
		return map[string]string{}
	}
	baseURL := cfg.BaseURL
	// for people who used previous versions
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	openai_apikey := cfg.APIKey
	fromLang := cfg.FromLang
	toLang := cfg.ToLang
	fmt.Println("\nGenerating...")

	for {
		url := baseURL

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

		content := fmt.Sprintf(promptTemplate, fromLang, toLang, toLang, fromLang, toLang)

		opai_req := OpenAIRequest{
			Model: model,
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
			fmt.Println("Error: The responce from LLM is incorrect.\nCheck your API key or network")
			return map[string]string{}
		}

		// contentを取り出す
		result := openAIResp.Choices[0].Message.Content

		// fmt.Println(result)
		var fields map[string]string
		json.Unmarshal([]byte(result), &fields)

		fmt.Println("\nFront         :", fields["Front"])
		fmt.Println("Back          :", fields["Back"])
		fmt.Println("Front Sentence:", fields["Front_Sentence"])
		fmt.Println("Back Sentence :", fields["Back_Sentence"], "\n")

		if fields != nil {
			return fields
		}
	}
}
