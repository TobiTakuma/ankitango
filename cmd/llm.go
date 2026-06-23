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
  Keep it 6–9 words. Avoid overly simple drills and overly long multi-clause sentences.

- "Back_Sentence": the translation of that sentence in %s, using the SAME sense as Front_Sentence

No markdown, no explanation, only the JSON object.

Example of a word with multiple senses, "patient":
{"Front":"patient","Back":"患者；忍耐強い","Front_Sentence":"The doctor examined the patient carefully.","Back_Sentence":"医者は患者を注意深く診察した。"}

Example of a single-sense word, "apple":
{"Front":"apple","Back":"りんご","Front_Sentence":"He bought fresh apples at the market.","Back_Sentence":"彼は市場で新鮮なリンゴを買った。"}

Example of a word that looks multi-sense but is ONE sense, "hello" (greeting only, do NOT split by situation):
{"Front":"hello","Back":"こんにちは","Front_Sentence":"She said hello to everyone she met.","Back_Sentence":"彼女は会う人みんなに挨拶した。"}`

// generate new words
func generateWord(word string) (map[string]string, error) {
	cfg := loadConfig()
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("No API key has been configured.\nPlease configure it using `ankitango config apikey <provider> <key>`.")
	}
	if cfg.FromLang == "" || cfg.ToLang == "" {
		return nil, fmt.Errorf("No language has been configured.\nPlease configure it using `ankitango config lang <from> <to>`.\nExample: `ankitango config lang English Japanese`")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	model := cfg.LLMModel
	if model == "" {
		model = "gpt-4o-mini"
	}
	fromLang := cfg.FromLang
	toLang := cfg.ToLang

	for {
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

		req, _ := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Network error: %w", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var openAIResp OpenAIResponse
		json.Unmarshal(body, &openAIResp)

		if len(openAIResp.Choices) == 0 {
			return nil, fmt.Errorf("Invalid response from LLM. Check your API key or network.")
		}

		result := openAIResp.Choices[0].Message.Content
		var fields map[string]string
		json.Unmarshal([]byte(result), &fields)

		if fields != nil {
			return fields, nil
		}
	}
}
func printLLMresult(fields map[string]string) {
	fmt.Println("\nFront         :", fields["Front"])
	fmt.Println("Back          :", fields["Back"])
	fmt.Println("Front Sentence:", fields["Front_Sentence"])
	fmt.Println("Back Sentence :", fields["Back_Sentence"], "\n")
}
