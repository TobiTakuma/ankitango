package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Config struct {
	APIKey      string `json:"api_key"`
	FromLang    string `json:"fromlang"`
	ToLang      string `json:"tolang"`
	BaseURL     string `json:"base_url"`
	LLMModel    string `json:"model"`
	tuiDeckName string `json:"tuiDeck"`
}

type Provider struct {
	BaseURL string
	Model   string
}

var providers = map[string]Provider{
	"openai": {"https://api.openai.com/v1/chat/completions", "gpt-4o-mini"},
	"gemini": {"https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", "gemini-2.5-flash"}}

var home, _ = os.UserHomeDir()
var configPATH = home + "/.config/ankitango/config.json"

func loadConfig() Config {
	data, err := os.ReadFile(configPATH)
	if err != nil {
		return Config{}
	}

	var cfg Config
	json.Unmarshal(data, &cfg)
	return cfg
}

func saveConfig(cfg Config) {
	os.MkdirAll(home+"/.config/ankitango", 0755)
	data, _ := json.Marshal(cfg)
	os.WriteFile(configPATH, data, 0644)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage settings (apikey / lang / show)",
}

var configApiKeyCmd = &cobra.Command{
	Use:   "apikey [provider] [key] [model]",
	Short: "Save the API key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: specify provider and key (Ex: ankitango apikey openai sk-...)")
			return
		}

		p, ok := providers[args[0]]
		if !ok {
			fmt.Println("Error: unknown provider: ", args[0])
			fmt.Println("\nProvider List")
			for name := range providers {
				fmt.Println(name)
			}
			return
		}

		cfg := loadConfig() // read current setting
		cfg.BaseURL = p.BaseURL
		cfg.LLMModel = p.Model
		if len(args) >= 3 {
			cfg.LLMModel = args[2]
		}
		cfg.APIKey = args[1] // Override API key
		saveConfig(cfg)
		fmt.Printf("Provider: %s, key saved!\n", args[0])
	},
}

var configLangCmd = &cobra.Command{
	Use:   "lang [fromLang][toLang]",
	Short: "Save the language settings",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("Error: specify language")
			return
		}

		cfg := loadConfig()
		cfg.FromLang = args[0]
		cfg.ToLang = args[1]
		saveConfig(cfg)

		fmt.Printf("Save language setting: %s to %s\n", args[0], args[1])
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "View current settings",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		key := cfg.APIKey
		model := cfg.LLMModel
		if len(key) > 10 {
			key = key[:3] + "..." + key[len(key)-4:]
		}

		fmt.Println("LLM Model: " + model)
		fmt.Println("APIkey: " + key)
		langConfig := fmt.Sprintf(`Lang: "%s" to "%s"`, cfg.FromLang, cfg.ToLang)
		fmt.Println(langConfig)

	},
}
