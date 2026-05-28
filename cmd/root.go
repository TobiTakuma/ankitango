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
	rootCmd.AddCommand(listCmd)
}
