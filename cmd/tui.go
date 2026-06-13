package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// ============================================================================
// bubbletea の考え方（最初にここを読む）
// ----------------------------------------------------------------------------
// bubbletea は起動すると、裏で次のループをずっと回し続ける：
//
//   Init → ┌──────────────────────────────────────┐
//          │ 何か起きる（キー入力など）= Msg が届く │
//          │              ↓                         │
//          │ Update が呼ばれる → 状態(model)を更新  │
//          │              ↓                         │
//          │ View が呼ばれる → model を画面に描く   │
//          └───────────────┬──────────────────────┘
//                          └─→ 最初に戻る（Quit するまで）
//
// 大事なこと：
//   ・Update も View も "自分では呼ばない"。bubbletea が勝手に呼ぶ。
//     自分が書くのは「呼ばれたら何を返すか」だけ。
//   ・状態(model)が変わるのは Update の中だけ。
//     View は model を見て描くだけで、何も変えない。
//
// 機能を足したくなったら、毎回この3つを自問する：
//   1. どんな状態が要る？      → model にフィールドを足す
//   2. どの Msg で、どう変える？ → Update に case を足す
//   3. その状態をどう描く？    → View に表示を足す
// ============================================================================

// tui コマンド。`ankitango tui` で起動する
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Run: func(cmd *cobra.Command, args []string) {
		// initialModel() で作った初期状態を渡してプログラムを作る。
		// p.Run() がさっきのループを開始し、Quit されるまで戻ってこない。
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error: could not start TUI:", err)
		}
	},
}

// ----- bubbletea の3点セット（Model / Update / View）-----

// Model = 画面の状態を入れる箱。「今この瞬間の画面の全データ」をここに持つ。
// 画面に出ているものは、必ずこの中のどれかに対応している。
// 画面を増やしたくなったら、まずここにフィールドを足す（例: step int で今どの画面か）。
type model struct {
	input textinput.Model // 入力欄。これ自体が小さな bubbletea 部品（bubbles）
	word  string          // Enter で確定した単語。確定前は "" のまま
}

// initialModel = 起動時の状態を1個作って返す。ここが「画面の初期値」。
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "type a word..." // 何も打っていない時に薄く出る案内文
	ti.Focus()                        // カーソルをこの入力欄に置く（複数欄ある時は1つだけ Focus する）

	return model{input: ti}
}

// Init = 起動直後に1回だけ呼ばれる。「最初にやっておきたい仕事」を tea.Cmd で返す。
// ここでは textinput.Blink を返して、カーソルの点滅を始めている。
// （特に何もしたくない時は return nil でよい）
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update = 何か起きるたびに呼ばれる。引数 msg に「何が起きたか」が入っている。
// 戻り値は (新しい状態, 次にやらせたい仕事)。状態を変えるのはこの中だけ。
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// msg は「キー入力」「ウィンドウサイズ変更」など色々な型が来る。
	// type switch で「今回はどの種類の Msg か」を判定する。
	switch msg := msg.(type) {

	// tea.KeyMsg = キーが押された、という Msg
	case tea.KeyMsg:
		// msg.String() で押されたキーが文字列で取れる（"a" "enter" "ctrl+c" など）
		switch msg.String() {
		case "ctrl+c", "esc":
			// tea.Quit を返すと、あの無限ループが終わってプログラムが終了する
			return m, tea.Quit

		case "enter":
			// 入力欄に今入っている文字列を取り出して、確定値として state に保存する。
			// → 状態が変わったので、この後 bubbletea が View を呼び直し、画面に反映される。
			m.word = m.input.Value()
			// TODO: ここから先へ画面を進める（デッキ選択 → generateWord → addCard）。
			//       例: m に step を持たせ、ここで m.step = 1 にして View を分岐させる。
			return m, nil
		}
	}

	// 上の case に当てはまらなかった入力（普通の文字・矢印キー・Backspace など）は、
	// 入力欄の部品自身に処理してもらう。input.Update が「文字が増えた入力欄」を返すので、
	// それで m.input を差し替える。cmd は入力欄が必要とする内部の仕事（点滅など）。
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View = 今の状態(model)を1本の文字列に変換するだけ。画面に出るのはこの戻り値。
// ここでは絶対に状態を変えない（変えるのは Update の仕事）。
// 画面を切り替えたい時は、ここで m の中身を見て if で出し分ける。
func (m model) View() string {
	s := "Add a word to Anki\n\n"
	s += m.input.View() + "\n\n" // 入力欄も「自分を文字列にする View」を持っている
	if m.word != "" {
		// 確定済みなら、確定した単語も表示する
		s += "word: " + m.word + "\n\n"
	}
	s += "(enter: confirm / esc: quit)\n" // 操作のヒント（フッター）
	return s
}
