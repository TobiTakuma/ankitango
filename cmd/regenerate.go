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

		if !checkAnkiRunning() {
			return
		}
		if !isDeck(getDeckName(), deckName) {
			return
		}
		if !IsModel() {
			addNewModel()
		}

		var noteIds []int
		noteIds = findNotes(deckName)

		for i := 0; i < len(noteIds); i++ {
			fmt.Println("--------------------")
			var frontWord string
			frontWord = noteInfo(noteIds[i])

			fields, err := generateWord(frontWord)
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}
			printLLMresult(fields)
			updateNoteFields(noteIds[i], fields)
		}
	},
}
