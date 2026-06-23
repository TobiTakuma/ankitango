package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var regen = &cobra.Command{
	Use:   "regen [deckName]",
	Short: "Regenerate translations for all cards in a deck",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Error: Please specify the deckname.")
			return
		}
		var deckName = args[0]

		if err := checkAnkiRunning(); err != nil {
			fmt.Println("Error:", err)
			return
		}
		decks, err := getDeckName()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if err := isDeck(decks, deckName); err != nil {
			fmt.Println("Error:", err)
			return
		}
		modelExists, err := IsModel()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if !modelExists {
			if err := addNewModel(); err != nil {
				fmt.Println("Error:", err)
				return
			}
		}

		noteIds, err := findNotes(deckName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		for i := 0; i < len(noteIds); i++ {
			fmt.Println("--------------------")
			frontWord, err := noteInfo(noteIds[i])
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			fields, err := generateWord(frontWord)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printLLMresult(fields)
			if err := updateNoteFields(noteIds[i], fields); err != nil {
				fmt.Println("Error:", err)
				continue
			}
		}
	},
}
