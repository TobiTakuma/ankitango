package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type Config struct {
	APIKey   string `json:"api_key"`
	FromLang string `json:"fromlang"`
	ToLang   string `json:"tolang"`
}

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
	Short: "Manage the config data. It has three sub command",
}

var configApiKeyCmd = &cobra.Command{
	Use:   "apikey [key]",
	Short: "Save the api key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Error: specify API Key")
			return
		}

		cfg := loadConfig()  // read current setting
		cfg.APIKey = args[0] // Override API key
		saveConfig(cfg)
		fmt.Println("APIKey saved!")
	},
}

var configLangCmd = &cobra.Command{
	Use:   "lang [fromLang][toLang]",
	Short: "Save the langage settings",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
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
	Short: "View current setting",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := loadConfig()
		key := cfg.APIKey
		if len(key) > 10 {
			key = key[:3] + "..." + key[len(key)-4:]
		}
		fmt.Println("APIkey: " + key)
		langConfig := fmt.Sprintf(`Lang: "%s" to "%s"`, cfg.FromLang, cfg.ToLang)
		fmt.Println(langConfig)

	},
}
