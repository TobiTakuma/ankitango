package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// tui コマンド。`ankitango tui` で起動する
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Run: func(cmd *cobra.Command, args []string) {
		// bubbletea のプログラムを作って実行する
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error: could not start TUI:", err)
		}
	},
}

// ----- bubbletea の3点セット（Model / Update / View）-----

// Model = 画面の状態。今は「単語の入力欄」と「確定した単語」だけ
type model struct {
	input textinput.Model // 入力欄（bubbles の部品）
	word  string          // Enter で確定した単語
}

// 起動時の初期状態を作る
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "type a word..."
	ti.Focus() // カーソルを入力欄に置く

	return model{input: ti}
}

// Init = 起動直後に1回走る。カーソルの点滅を開始する
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update = キー入力などが来るたびに呼ばれ、新しい状態を返す
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			// 終了する
			return m, tea.Quit
		case "enter":
			// 入力された単語を確定する
			m.word = m.input.Value()
			// TODO: ここから先へ画面を進める（デッキ選択 → generateWord → addCard）
			return m, nil
		}
	}

	// それ以外のキーは入力欄に渡す（文字入力・カーソル移動など）
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View = 状態を文字列にして画面に描く
func (m model) View() string {
	s := "Add a word to Anki\n\n"
	s += m.input.View() + "\n\n"
	if m.word != "" {
		s += "word: " + m.word + "\n\n"
	}
	s += "(enter: confirm / esc: quit)\n"
	return s
}
