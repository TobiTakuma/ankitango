package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all decks in Anki",
	Run: func(cmd *cobra.Command, args []string) {
		if !checkAnkiRunning() {
			return
		}

		printList(getDeckName())
	},
}

func printList(list []string) {
	for i := 0; i < len(list); i++ {
		fmt.Println(list[i])
	}
}
