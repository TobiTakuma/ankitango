package cmd

import (
	"github.com/spf13/cobra"
)

// definition of "AddAnki" command
var rootCmd = &cobra.Command{
	Use:   "ankitango",
	Short: "CLI tool that generate word and sentence then add to Anki aoutmatically",
}

// the entrance of main.go find the command
func Execute() {
	rootCmd.Execute()
}

// connect addCmd to root command
func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().StringP("file", "f", "", "words file(.txt/.csv)")

	rootCmd.AddCommand(listCmd)

	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configApiKeyCmd)
	configCmd.AddCommand(configLangCmd)
	configCmd.AddCommand(configShowCmd)

	rootCmd.AddCommand(tuiCmd)
}
