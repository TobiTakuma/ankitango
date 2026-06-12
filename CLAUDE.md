# Anki自動追加アプリ CLAUDE.md

## 完成までの計画

基本機能・配布は一通り動く状態だが、今後も追加機能を実装していくため**プロジェクト自体は未完成**（継続開発中）。

| フェーズ | 内容 | 状態 |
|---------|------|------|
| 1. AnkiConnect基本機能 | `checkAnkiRunning()`, `getDeckName()`, `addNote()` | ✅ 完了 |
| 2. OpenAI API連携 | 単語を入力 → 翻訳・例文を生成 | ✅ 完了 |
| 3. 機能追加 | デッキ・モデル選択、重複チェック、カード方向管理、config設定 | ✅ 完了 |
| 4. CLIコマンド化 | cobra導入、`ankitango add <word>` などのコマンド実装 | ✅ 完了 |
| 5. 配布 | GoReleaserでリリース、install.sh / install.ps1 作成 | ✅ 完了 |
| 6. ファイル一括取り込み | `add -f <file>` で `.txt`/`.csv` から複数単語を追加 | ✅ 完了 |
| 7. TUI | bubbletea導入、画面フロー実装（AIに任せる） | 未着手 |
| 8. 追加機能（継続） | 自動判別・複数訳対応・発音/音声フィールド・他LLM対応など | 🚧 検討/実装中 |

## 後でやること

- **【接続切れ対策】残タスク**（2026-06-04時点。コア機能は完成済み）
  - `generateWord`（OpenAI側）の `panic` をリトライ＋エラー返しに（Anki側と別対応。429レート制限はこちらで起きうる）。生成失敗の単語も `failedWords` に記録するか合わせて検討
  - `getDeckName` / `IsModel` / `addNewModel` の `panic` をエラー表示に（ループ外で1回だけ呼ばれるので優先度低）
- `panic(err)` をエラーメッセージ表示に変更（接続切れ対策で `ankiInvoke`＋`addCard`＋`isNote` は対応済み。`getDeckName` / `IsModel` / `addNewModel` / `generateWord` はまだ `panic` のまま）
- 発音記号・音声データフィールドの追加（`Pronunciation`, `Audio` フィールドは存在するがAIが生成していない。生成内容に含める）
- 複数の意味・訳語への対応（1回の実行で複数カードを生成するか検討）
- 拡張子での自動判別（`-f` なしで `add words.txt MyDeck` と打ったら、args[0] が `.txt`/`.csv` なら自動でファイル扱いにする）：現状は `-f` フラグ必須。優先度低
- 他LLM対応：OpenAI / Gemini は **OpenAI互換エンドポイント方式**で対応完了（2026-06-05）。残り：
  - ローカルモデル（Ollama/LM Studio等）：プリセットに1行足すだけで可能だが「あとでやる」。`generateWord` の「APIキー空ならエラー」チェックに引っかかるので、ダミーキー（`config apikey local ...`）かチェック緩和が必要
  - Claude：`providers` マップに `claude` 行を足すだけで対応可（互換エンドポイント）
  - 必要なら「プロバイダごとにキー保存」（切り替えのたびにキー再入力が面倒になったら）
  - `generateWord` 末尾の DEBUG 出力（`fmt.Println("DEBUG ...")`）を消したか確認

---

## プロジェクト構成・セットアップ・使い方

### 全体像（現在の構成）
```
main.go              ← エントリーポイント。cmd.Execute() を呼ぶだけ
cmd/
  root.go            ← ルートコマンド（ankitango）の定義。init() で各コマンド・フラグを登録
  add.go             ← add / list コマンド。AnkiConnect・OpenAI関連の関数群もここ
                        （addCmd, listCmd, generateWord, addCard, isNote, isDeck,
                         readWord, IsModel, addNewModel, getDeckName, checkAnkiRunning）
  config.go          ← config コマンド（apikey / lang / show）と loadConfig / saveConfig
sample/              ← テスト用 .txt / .csv（.gitignore 済み、コミットしない）
install.sh           ← Mac/Linux 用インストールスクリプト
install.ps1          ← Windows 用インストールスクリプト
.goreleaser.yaml     ← GoReleaser 設定
.github/workflows/release.yml ← タグpushで自動リリース
```
- モジュール名: `github.com/TobiTakuma/ankitango`
- 設定ファイル: `~/.config/ankitango/config.json`（APIキー・言語）
- 自作Ankiモデル名: `ankitango`（初回実行時に自動作成）

### セットアップ手順（インストール）
**macOS / Linux**（curl でインストール）
```bash
curl -fsSL https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.sh | sh
```
**Windows**（コマンドからインストール）
```powershell
powershell -Command "Invoke-WebRequest -Uri https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.ps1 -OutFile install.ps1; .\install.ps1"
```
**Go環境がある場合**
```bash
go install github.com/TobiTakuma/ankitango@latest
```
インストール後に設定:
```bash
ankitango config apikey <key>             # APIキーを設定
ankitango config lang English Japanese    # 言語を設定（from to）
ankitango config show                     # 現在の設定を表示
```
※ Anki を起動し、AnkiConnect アドオンを入れておくこと。

### 使い方（現在）
```bash
ankitango add apple MyDeck                # 単語1個を追加
ankitango add "look up" "My Deck"         # スペースを含む単語/デッキはクォートで囲む
ankitango add -f sample/words.txt MyDeck  # .txt から複数単語を一括追加（1行1単語）
ankitango add -f sample/words.csv MyDeck  # .csv から一括追加（1セル1単語、列数バラバラOK）
ankitango list                            # デッキ一覧を表示
ankitango config apikey <key>             # APIキーを設定
ankitango config lang <fromLang> <toLang> # 言語を設定（例: English Japanese）
ankitango config show                     # 現在の設定を表示
```
- `-f`（`--file`）は string型フラグなので「`-f <ファイル名>`」と値をセットで渡す
- ファイル取り込み時、重複・生成失敗の単語は `continue` でスキップし残りを処理する

---

## 変更ログ

### 2026-05-25
**Claude:**
- `python/` フォルダを作成し `main.py`, `app.py`, `test.py` を移動（参照用）
- `go.mod` を初期化（モジュール名: `github.com/TobiTakuma/add_anki`）
- `main.go`, `cmd/`, `internal/` を作成

**ユーザーの指示により:**
- `cmd/`, `internal/` を削除。当面は `main.go` 1ファイルに書く方針に変更

### 2026-05-26
**ユーザー:**
- `getDeckName()`: デッキ一覧を取得してprintする
- `addNote()`: カードをAnkiに追加する（deckName, modelName, fields, tags）
- `checkAnkiRunning()`: Ankiの起動確認。未起動なら `false` を返してメッセージを表示
- `generateWard()`: OpenAI APIを呼び出して単語を翻訳する関数を実装・動作確認済み
- `addNewModel()`: AnkiConnectでカスタムモデル（Front/Front_Sentence/Back/Back_Sentence）を作成・動作確認済み
- `IsModel()`: モデルが存在するか確認してboolを返す関数を実装
- `generateWard(word, fromLang, toLang)`: 言語を引数で受け取り4フィールドのJSONを返すように修正

### 2026-05-27
**ユーザー:**
- プロジェクトを `add_anki/` から `ankitango/` に移動・リネーム
- cobra導入。`cmd/add.go`, `cmd/root.go`, `main.go` の構成に変更
- `add` コマンド実装：`ankitango add [word] [deckName]`
- `generateWard()` のリトライ処理を実装（JSONでなければ無限ループで再試行）
- `addCard(fields, deckName)` にdeckNameを引数として追加

### 2026-05-28
**ユーザー:**
- `list` コマンド実装：`ankitango list` でデッキ一覧を表示
- `isDeck()` 関数追加：存在しないデッキ名を指定した時にエラーメッセージを表示
- `getDeckName()` を `[]string` を返す関数に変更
- `addCard()` のレスポンス型を `*int64` に修正、成功メッセージを追加
- `isNote()` 関数追加：generateWard前に重複チェック。`findNotes` アクションを使用
- `godotenv` を削除し、`os.Getenv()` のみで環境変数からAPIキーを取得するように変更
- 不要ファイル（`.venv`, `.idea`, `python/`）を整理
- `go install .` でローカルバイナリとして使えるように設定。`~/.zshrc` に `export PATH=$PATH:$HOME/go/bin` を追加
- git push時にAPIキー漏洩エラー（`test.py` に古いキーが残っていた）。`.git` を削除して `git init` からやり直し
- `cmd/config.go` を実装：`Config` struct、`loadConfig()`、`saveConfig()`、`configCmd`・`configApiKeyCmd`・`configLangCmd`・`configShowCmd` の定義と `Run` を実装
- `root.go` の `init()` に `configCmd` とサブコマンドを追加

**Claude:**
- README.md を書き直し（Goプロジェクト用に更新）
- README.md・CLAUDE.md を更新（`go install` でのセットアップ、configコマンドを追記）
- 変更ログと質問ログが混在していたのを修正。同日エントリを1つにまとめる形式に変更

### 2026-05-29
**ユーザー:**
- `generateWord()` で `loadConfig()` を呼び出し、APIキー・言語設定をconfigから取得するように変更。言語を変えての動作確認済み
- エラー処理を追加：APIキー未設定、言語未設定、OpenAI応答が空の場合、引数不足
- `InOrderFields` に `Pronunciation`, `Audio`, `Synonym`, `Note` を追加
- `.github/workflows/release.yml` を作成し GitHub Actions を設定
- `install.sh`（Mac/Linux用）・`install.ps1`（Windows用）インストールスクリプトを作成
- v1.0.0 をリリース
- v1.0.1　をリリース

**Claude:**
- README.md をインストールスクリプトの手順を含む英語版に書き直し
- CLAUDE.md の変更ログ・計画表・質問ログを更新

### 2026-05-31
**ユーザー:**
- `add` コマンドで `.txt`/`.csv` ファイルから複数単語を一括取り込む機能を実装・完成
- `root.go` の `init()` に `addCmd.Flags().StringP("file", "f", "", ...)` を追加（`-f` フラグ登録）
- `add.go` の `Run` を書き換え：`filePath` の有無でファイルモード/単語モードに分岐し、`wordsArray` をループ処理。失敗時は `return` でなく `continue` でスキップ
- `add.go` の `Use` を `add [words] [deckName]` に変更
- `readWord(path)` 実装：`os.Open` → 拡張子で分岐（`.csv` は `encoding/csv`、`.txt` は `bufio.Scanner`）→ `strings.TrimSpace` で空行スキップ → `[]string` を返す
- CSV読み込みで `reader.FieldsPerRecord = -1` を設定し、行ごとにセル数がバラバラでも読めるように
- 引数チェック追加：ファイルモードは `len(args) < 1`、単語モードは `len(args) < 2`（panic防止）
- `sample/` を作成し `words.txt`/`words.csv` をテスト用に追加。`.gitignore` に `sample/` を追加
- デバッグ用出力・エラーメッセージの文言を整理（ファイルモードは「deckname のみ」に修正）

**Claude:**
- ファイル取り込み機能の設計・各コードの意味を説明（実装はユーザー本人が担当）
- `sample/` ディレクトリとテスト用 `.txt`/`.csv` を作成、`.gitignore` に追加
- `go build` / `go vet` で全体の動作を確認（全パターンOK）
- CLAUDE.md を全面更新：計画表（継続開発中に変更・ファイル取り込みフェーズ追加）、後でやること（完了分を削除）、構成/セットアップ/使い方を現状に合わせて書き直し

### 2026-06-04（連続リクエスト時の接続切れ対策・コア機能完成）
**ユーザー:**
- 連続リクエスト中に一瞬Ankiとの接続が切れて全体が死ぬ問題を解決
- `ankiInvoke(req any) ([]byte, error)` 共通関数を作成：`json.Marshal` → `http.Post` を最大3回リトライ（`for i := 0; i < 3`、失敗時 `time.Sleep(500ms)`）→ 成功で `resp` に代入し `break` → 3回ダメなら `resp == nil` で `error` を返す → body を返す
- AnkiConnect系の5関数（`isNote` / `IsModel` / `addNewModel` / `addCard` / `getDeckName`）を `ankiInvoke(req)` 呼び出しに置き換え（重複していたMarshal→POST処理を共通化）
- `checkAnkiRunning` は門番なので直接 `http.Post` のまま残す（リトライ不要、即「未起動」判定したいため）
- `addCard` を `error` を返す形に変更（接続失敗→`return err`、Anki側拒否→`fmt.Errorf("anki error: %s", ...)`、成功→`return nil`）
- `isNote` を `(bool, error)` を返す形に変更（接続失敗を握りつぶさず、ループ側で `failedWords` に記録できるように）
- `Run` のループで `failedWords []string` をローカルに用意し、`isNote`/`addCard` 失敗時に `append` して `continue`（失敗単語の収集）。エラー表示も関数内でなくループ側に集約し、単語付きで1箇所表示に統一
- `fail(words []string, path string)` を実装：`os.Create(path)` でファイルを作り（上書き方式）、`fmt.Fprintln` で1行1単語ずつ書き出す
- `fail.txt` の保存先を入力ファイルと同じフォルダに：`filepath.Dir(filePath)` + `filepath.Join(..., "fail.txt")`。ファイルモード（`filePath != ""`）かつ失敗があるときだけ書き出す
- 動作確認：途中でAnkiを終了させても panic せず最後まで回り、失敗単語が `fail.txt` に1行1単語で記録されることを確認（`-f` で再投入可能な形式）

**Claude:**
- 原因の見立てを説明：レート制限ではなくGoのHTTP keep-alive接続の使い回しが原因（後述の実装ポイント参照）。リトライが効く理由・`DisableKeepAlives` という別解も提示
- リトライの書き方、`resp` のスコープ問題、`ankiInvoke` 共通化の意味、`error` 戻り値の設計、`(bool, error)` の片方を `_` で捨てる書き方、失敗単語をどこで持つか（グローバルでなくループ内ローカル）、`fail` のファイル書き込み（`os.Create`/`fmt.Fprintln`）、`filepath.Dir`/`filepath.Join` での保存先指定を段階的に説明（実装はユーザー本人）
- レアなEOFの再現を待つより「わざと失敗を注入してコードの通り道を全部踏む」検証方法を提案（Anki終了・ポート変更）
- エラー表記の整理（typo・単語抜け・表記揺れ）は「関数内でprintせずループ側1箇所に集約」で根本解決、`fail.txt` に見出し行を混ぜない注意を説明
- `go build` / `go vet` で確認。残課題（`generateWord` のpanic等）を「後でやること」に反映

### 2026-06-04（重複検知スコープの修正・案B採用）
**ユーザー:**
- 別デッキからエクスポートした単語を新デッキに取り込む際、一部の単語が `isNote` をすり抜けて `addCard` で重複エラーになる問題に対応
- **案B採用**：`addCard`（addNote）に `options` を追加し `duplicateScope: "deck"` を指定。重複チェックを「コレクション全体」から「ターゲットデッキ内のみ」に変更（`Note` struct に `Options{ AllowDuplicate, DuplicateScope }` を追加）→ 他デッキに同じ単語があってもターゲットデッキに無ければ追加できるように
- エラー表示を改善：接続失敗（再挑戦すべき）と Anki側拒否＝重複（再挑戦しても無駄）を区別。`ankiInvoke` の接続失敗を sentinel error `ErrConnection`（`errors.New`）で返し、ループ側で `errors.Is(err, ErrConnection)` で判定。接続失敗のみ `fail.txt` に記録、重複は表示だけでスキップ（`fail.txt` に恒久的に失敗するゴミが溜まるのを防ぐ）

**Claude:**
- isNote通過→addCard重複の原因を説明：①Ankiの重複判定は同じノートタイプ内のみ（デッキAの元カードは別の型なので ankitango の追加を邪魔しない＝大半が成功する理由）②`isNote` はデッキ絞り込み・addNote はコレクション全体とスコープがズレていた ③`isNote` は入力した生の単語で検索するが、実際に保存される Front は AI が変換した語（例：「I'm curious about」→「curious」）なので、AIが語を変えると事前チェックがすり抜ける
- 案A（isNoteを全体検索に）と案B（addNoteをデッキ単位に）を提示。移行が目的なら案Bを推奨
- sentinel error + `errors.Is` による失敗種別の区別方法を説明

### 2026-06-05（複数LLM対応・OpenAI / Gemini）
**ユーザー:**
- OpenAI に加え Gemji（と将来ローカルモデル）に対応。LiteLLM は Python製/プロキシ常駐が必要で単一バイナリCLIに不向きと判断し、**OpenAI互換エンドポイント方式**を採用（送信コード・JSON構造はそのまま、baseURL・model・key だけ差し替え）
- config.go：`Config` に `BaseURL`（`json:"base_url"`）/ `Model`（`json:"model"`）を追加。`Provider{ BaseURL, Model }` 型と `providers` プリセットマップ（`openai` / `gemini`）を定義
- `config apikey` コマンドを `[provider] [key] [model]` 形式に変更：プリセットから baseURL/model を自動セット＋キー保存。未知のプロバイダは `range` で一覧表示。第3引数でモデルを上書き可能にし、**モデルが変わっても再ビルド不要**に
- add.go：`generateWord` の直書き url / `Model: "gpt-4o-mini"` を `cfg.BaseURL` / `cfg.Model` に差し替え（空ならOpenAIをデフォルトにフォールバック）。JSON入出力・`OpenAIRequest` 構造体は変更不要（互換なので同じ形）
- gemini プリセットのデフォルトモデルを `gemini-2.0-flash` → `gemini-2.5-flash` に変更（2.0-flash は当該アカウントで無料枠 `limit: 0` だったため）
- 動作確認：OpenAI（デフォルト）OK、Gemini（gemini-2.5-flash）OK

**Claude:**
- LiteLLM が Go/単一バイナリに不向きな理由と、OpenAI互換エンドポイント（OpenAI/Gemini/ローカルすべて同形式・Bearer認証）で代替できることを説明
- 「一番シンプルな方法＝ baseURL/model/key の3つだけ変える」設計、プリセットマップで持つ拡張性（1行追加で新プロバイダ）、番号より名前推奨を提案
- `:-`（→`:=`）のtypo、comma-okイディオム、map を `range` で回す書き方、`Config` にフィールド未追加によるエラーを指摘
- 429 RESOURCE_EXHAUSTED の `limit: 0` は「使い切った」でなく「そのモデルの無料枠が元々ゼロ」と説明。コードは正常でアカウント側の問題と切り分け、別モデル/課金/キー確認を提案
- モデル名を引数で上書き可能にして再ビルド不要にする設計を提案
- README.md を更新：Requirements（OpenAI/Gemini）、`config apikey <provider> <key> [model]` の書き方とGemini例・モデル上書き例、`add -f` でのファイル一括取り込みの使い方と `fail.txt` の説明を追記。さらに「Supported LLMs」セクションを追加（対応プロバイダ表・デフォルトモデル・キー取得先・モデル上書き・不明プロバイダ実行で一覧表示）を将来追加前提で記載

### 2026-06-12（TUI着手・最小雛形をAIが作成）
**ユーザー:**
- TUI対応の進め方を「最初だけAI、あとは自分」に決定（雛形はAI、画面追加は本人）

**Claude:**
- 使用ライブラリを説明：bubbletea（Elm Architecture の Model/Update/View でTUIを作るフレームワーク）、bubbles（textinput等の既製部品）、lipgloss（スタイリング）。`go get` で3つ追加し `go mod tidy`
- `cmd/tui.go` を新規作成：`tuiCmd`（`ankitango tui` で起動）＋ 最小の `model`/`initialModel`/`Init`/`Update`/`View`。単語入力欄1つ・Enterで確定・esc/ctrl+cで終了。`Update` の Enter 部分に「デッキ選択→generateWord→addCard へ進める」TODOコメントを残した
- `root.go` の `init()` に `rootCmd.AddCommand(tuiCmd)` を追加
- 既存ロジック（`getDeckName`/`generateWord`/`addCard`）はTUIからそのまま呼べる（TUIは見た目の層）。`go build`/`go vet` OK
- 次の一歩：`add` の流れ（単語入力→デッキ選択→AI生成プレビュー→確認して追加）を画面（state）として足していく

---

## 実装のポイント（学んだこと）

- AnkiConnectはHTTP POSTでJSONを送るだけ。言語は関係ない
- GoのJSON扱い: `json.Marshal()` でGoの値→JSONバイト列、`json.Unmarshal()` で逆変換
- `http.Post(url, "application/json", bytes.NewBuffer(jsonData))` の形で送る
- `resp.Body` は必ず `defer resp.Body.Close()` で閉じる
- AnkiConnectのレスポンスは `{"result": ..., "error": ...}` の形
- 連続リクエストで出る `Post EOF` の正体：Goは一度開いたTCP接続を次のリクエストで使い回す（HTTP keep-alive）。AnkiConnect側が先にアイドル接続を閉じると、Goが「死んだ接続」を再利用した瞬間に `EOF` になる。単発では起きず連続実行で確率が上がるタイミング依存の競合状態。レート制限（クールタイム）ではない。対策はリトライ（新しい接続を開き直すので直る）＋必要なら `&http.Transport{DisableKeepAlives: true}` で使い回し自体を止める
- 失敗の2種類を区別する：`http.Post` の `err`（通信失敗・EOF）はリトライで直る／レスポンスの `ankiResp.Error != nil`（重複・デッキ無しなど）は何度送っても同じなのでリトライせず即失敗扱い
- 失敗単語の集計はグローバル変数でなく `Run` のループ内ローカル `[]string` に持つ（「どの単語か」と「成否」を同時に知るのはループだけ。リセット不要・隠れた状態を避けられる。ファイル書き込みはループ後に1回だけ）
- `fmt.Errorf`/`Printf` は値を埋めるのに `%s` などの穴が必要（`Println` は穴なしで値を並べられる、の違い）
- Ankiの重複判定は**同じノートタイプ（モデル）の中だけ**で行われる。ankitango は専用モデル `ankitango` を使うので、他の型（Basic等）の同じ単語は重複扱いされない。AnkiConnect `addNote` の重複スコープは `options.duplicateScope` で制御：デフォルトはコレクション全体、`"deck"` を指定するとターゲットデッキ内のみ（`options` は `note` オブジェクトの中に置く）
- `isNote`（事前チェック）の盲点：検索は**入力した生の単語**で行うが、実際にカードへ保存される `Front` は **AIが変換した後の語**。両者がズレると（フレーズ→単語など）事前チェックをすり抜け、addNote 側で本物の重複として弾かれる
- 失敗を「再挑戦すべき（接続失敗）」と「再挑戦しても無駄（重複・拒否）」で分ける。sentinel error（`var ErrConnection = errors.New(...)`）を返し、呼び出し側で `errors.Is(err, ErrConnection)` で判定するのがGoの定番。前者だけ `fail.txt` に記録する
- 複数LLM対応は **OpenAI互換エンドポイント**が単一バイナリCLIに最適。OpenAI・Gemini・ローカル（Ollama/LM Studio等）はどれも `/chat/completions` に `Authorization: Bearer KEY` で OpenAI形式のJSONを送れる。違うのは baseURL・model・key の3つだけ＝送信コードを流用できる。LiteLLM は Python製/プロキシ常駐が必要で配布に不向き
  - Gemini互換: `https://generativelanguage.googleapis.com/v1beta/openai/chat/completions`／Claude互換: `https://api.anthropic.com/v1/chat/completions`（URLは変わりうるので要確認）
  - `response_format: json_object` は互換先で効かないことがあるが、`generateWord` の「JSONになるまで再試行」ループが吸収する
- 設定は `map[string]Provider` のプリセットで持つと新プロバイダ追加が1行で済む。よく変わるモデル名はコマンド引数で上書き可能にしておくと**再ビルド不要**で対応できる（baseURL はほぼ変わらないのでコード直書きで十分）
- Gemini の無料枠は**モデルごとに違う**。`429 RESOURCE_EXHAUSTED` の `limit: 0` は「使い切った」ではなく「そのモデルの無料枠が元々ゼロ」の意味（例: gemini-2.0-flash が limit:0 でも gemini-2.5-flash は使えた）。コードでなくアカウント/モデル側の問題
- Goの map は整数添字で取り出せない（`m[0]` 不可）。一覧は `for k := range m {}` で回す。`v, ok := m[key]` の comma-ok でキー存在を確認

---

## 質問ログ

### 2026-06-12
**質問：** TUIに対応させたい。どこから始めればいい？

**回答：**
- 使うのは bubbletea（TUIフレームワーク。Model=状態 / Update=入力で更新 / View=描画 の3点セット）／bubbles（textinput等の部品）／lipgloss（見た目）
- 始め方：①画面フローを紙で決める（`add` と同じ 単語入力→デッキ選択→AI生成プレビュー→確認）②`cmd/tui.go` に `tuiCmd` を作り `root.go` で登録 ③最小の Model/Update/View（入力欄1つ）から動かす ④既存関数（`getDeckName`/`generateWord`/`addCard`）はそのまま呼ぶ＝ロジックは流用、TUIは見た目の層
- いきなり全画面を作らず「入力欄1つ」から始めるのがコツ
- 進め方は「最初だけAI、あとは自分」を選択 → 最小雛形をAIが作成（変更ログ参照）

### 2026-06-09
**質問：** 全コマンドに `if !checkAnkiRunning() { return }` を書くのは無駄なので、`root.go` の `init()` に書けばよいのでは？ この推測は正しい？

**回答：**
- 「重複を1か所にまとめたい（DRY）」という方向性は正しいが、`init()` は不適切
- `init()` はパッケージ読み込み時に必ず1回走る。そこで `return` すると後続の `AddCommand` が実行されず、Anki停止時にコマンドが一切登録されない → `config apikey` すら「unknown command」になり、Anki無しでは設定もできなくなる（元より悪化）
- `init()` はどのコマンド（config・--help含む）でも走る。Anki が要るのは実際には `add` と `list` だけなので scope が広すぎる
- 正攻法は cobra の `PreRunE`：Anki が必要なコマンド（`addCmd`/`listCmd`）にだけ「`Run` 直前にチェック→失敗なら error を返す」フックを付ける。config には付けない。チェック関数を1つ作って両方の `PreRunE` に渡せば DRY も満たせる
- `init()` はコマンド・フラグ登録だけに戻すのが正しい

### 2026-06-05
**質問：** Gemini なども対応させたい。どうすればいい？

**回答：**
- 共通の窓口（interface）か、もっとシンプルには **OpenAI互換エンドポイント**を使う方法がある
- LLMごとに違うのは baseURL・APIキーの渡し方・リクエスト/レスポンスのJSON構造。共通なのはプロンプトと最終的な4フィールド変換
- 互換エンドポイントなら送るJSONが同じなので、baseURL・model・key を差し替えるだけで対応できる

### 2026-06-05
**質問：** LiteLLM が便利と聞いたが使えない？

**回答：**
- LiteLLM は Python のライブラリ/プロキシ。Goから直接は使えず、プロキシ常駐は単一バイナリ配布のCLIに不向き（ユーザーに別サーバーを要求してしまう）
- 同じ「1つの書き方で複数LLM」効果は **OpenAI互換エンドポイント**でサーバー無しに実現できる（OpenAI/Gemini/ローカルすべて対応）

### 2026-06-05
**質問：** 一番シンプルに OpenAI・Gemini（＋後でローカル）に対応するには？

**回答：**
- 全部 OpenAI形式・Bearer認証で叩けるので、プロバイダ分岐すら不要。`Config` に `BaseURL` / `Model` を足し、`generateWord` の直書き url/model をそれに差し替えるだけ
- 切り替えは baseURL・model・key の3つを設定するだけ。送信コードもリクエストJSONも流用

### 2026-06-05
**質問：** APIキーを入れる時に、どのモデルを使うかも同時に選ぶ方がいいのでは？（番号割り当てで `apikey 1 sk...` のように）

**回答：**
- 良い案。むしろ前案（URL直打ち）よりシンプル。プリセット表 `map[string]Provider` を持ち、`apikey [provider] [key]` で baseURL/model を自動セット＋キー保存にする
- 番号より名前（`openai`/`gemini`）推奨：自己説明的で、拡張時も表に1行足すだけ

### 2026-06-05
**質問：** `p, ok :- providers[...]` の `p` の所のエラーの意味は？

**回答：**
- `:-` は存在しない演算子。`:=`（宣言＋代入）のtypo。`p, ok := providers[args[0]]` が正しい（comma-okイディオム：`ok` はキーが存在すれば true）
- 同ブロックの一覧表示も map を整数で回せないので `for name := range providers {}` に直す

### 2026-06-05
**質問：** 71・72行目（`cfg.BaseURL` / `cfg.Model`）のエラーは？

**回答：**
- `Config` 構造体に `BaseURL` / `Model` フィールドをまだ追加していないため（`type Config has no field`）。手順①の2項目追加が抜けていた

### 2026-06-05
**質問：** url と model を add.go に反映するだけでいい？ JSONの書き方も変える必要ある？

**回答：**
- JSONは変更不要。`Config` にタグ付きフィールドを足したので `json.Marshal`/`Unmarshal` が自動で処理する（config.json に項目が自動で増える）
- 変えるのは `generateWord` の url と Model だけ。送るリクエストJSON（`OpenAIRequest`）も互換なのでそのまま

### 2026-06-05
**質問：** Gemini で `429` `limit: 0` が出た。これは？ また使えなくなったらコードを書き換えないといけない？

**回答：**
- `429 RESOURCE_EXHAUSTED` + `limit: 0` は「使い切った」ではなく「そのモデル（gemini-2.0-flash）の無料枠が元々ゼロ」。コードは正常（互換連携は成功）。別モデル（gemini-2.5-flash で解決）・課金有効化・キー確認で対応
- モデルが変わるたびにコードを書き換えたくないので、`apikey` の第3引数でモデルを上書き可能にしておく → 再ビルド不要でコマンドだけで変更できる

### 2026-06-04
**質問：** `isNote` では検知されないが `addCard` で重複エラーになる単語がある。特定のデッキでだけ起き、他のデッキでは検知される。なぜ？

**回答：**
- `isNote` はデッキ絞り込み検索（`deck:"X" Front:"word"`）、addNote はデフォルトでコレクション全体の重複チェック、とスコープがズレているのが原因
- その単語は別のデッキに存在する → isNote（デッキX限定）は見逃す → addNote（全体）が捕まえる、という流れ
- 直すには両方のスコープを揃える。案A（isNoteを全体検索に）か案B（addNoteをデッキ単位に）

### 2026-06-04
**質問：** 全単語が別デッキに存在するはずなのに、大半は重複問題が起きないのはなぜ？

**回答：**
- Ankiの重複判定は**同じノートタイプの中だけ**。デッキAの元カードは別の型（Basic等）なので、ankitango 型の追加では無視される＝大半が普通に追加できる
- 失敗する少数は、**過去のテストで ankitango 型として作った分**が他デッキ（テスト用）に残っているもの。isNote（ターゲットデッキ限定）は見逃し、addNote（ankitango型を全デッキ）が捕まえる
- 「全単語がAnki内にある」は本当だが、大半は別の型なので ankitango の重複チェックに引っかからない

### 2026-06-04
**質問：** 案B（duplicateScope: deck）にしたら1単語だけ失敗。「I'm curious about」→Front「curious」で `Connection failed` と出る。なぜ？

**回答：**
- まず表示が嘘：ループが addCard の失敗を理由を問わず「Connection failed」と決め打ち表示していた。本当の `err` を出すと重複エラーだった
- 本物の重複：入力「I'm curious about」をAIが「curious」に変換、その「curious」が既に WordHunt にあった＝正しく弾かれた（バグではない）
- isNoteが見逃した理由：isNoteは入力「I'm curious about」で検索、実際に保存される Front は「curious」でズレるため。AIが語を変えると事前チェックはすり抜ける
- 仕上げ：接続失敗（fail.txt行き）と重複（表示のみ）を `errors.Is(err, ErrConnection)` で区別する

### 2026-06-04
**質問：** `fail.txt` が今いるディレクトリに作られてしまう。読み込んだテキストファイルと同じ場所に置ける？

**回答：**
- できる。`filepath.Dir(filePath)` で入力ファイルのフォルダを取り出し、`filepath.Join(dir, "fail.txt")` で保存先パスを作って `fail` に渡す（`os.Create(path)`）
- `sample/test.txt` を読み込めば `sample/fail.txt` に出る。`filepath.Join` は `/` を自分で書かず、Windowsでも正しく動く
- 注意：単語モード（`filePath == ""`）だと `filepath.Dir("")` は `.`（カレント）になる。1単語の失敗のためにファイルを作るのは大げさなので、`filePath != ""`（ファイルモード）かつ失敗があるときだけ `fail` を呼ぶのがきれい

### 2026-06-04
**質問：** `fail` 関数の中身（ファイル書き込み）の作り方は？

**回答：**
- 「失敗単語の `[]string` を受け取って1行ずつファイルに書く」だけの関数にする
- 道具：`os.Create(path)`（なければ作成・あれば上書き。`readWord` の `os.Open`=読む の対）／`fmt.Fprintln(file, w)`（`Println` のファイル版。改行も自動）／`defer file.Close()`
- 上書き（`os.Create`）がおすすめ：`add -f fail.txt` で再挑戦するたびに、まだ失敗する単語だけが残り自然にリストが減る。追記だと重複が溜まる
- 呼び出しはループの後に `if len(failedWords) > 0 { fail(failedWords, path) }`（全成功時に空ファイルを作らない）
- 「failed words」などの見出しは画面表示だけにし、ファイルには1行1単語の純粋なリストのみ（`-f` 再投入時に見出しが単語として登録されるのを防ぐ）

### 2026-06-04
**質問：** `isNote` が `(bool, error)` を返すようになった。boolだけ／errorだけ取り出すには？

**回答：**
- Goは戻り値が2つなら受け取る変数も2つ必要。片方だけ書くとコンパイルエラー
- 両方使う：`exists, err := isNote(...)` ／ bool だけ：`exists, _ := isNote(...)` ／ error だけ：`_, err := isNote(...)`（`_` は「使わないので捨てる」記号。使わない変数を許さないGoの作法）
- ただし `isNote` を `(bool, error)` にした目的は接続失敗を握りつぶさないこと。捨てずに「error を先に確認 → 問題なければ bool を見る」が定番の流れ

### 2026-06-04
**質問：** 連続リクエストで途中で接続が切れて全部無に帰す問題を、①panicをやめてリトライ→クールタイムの有無を調べる ②エラーでも3回試す ③3回失敗した単語を.txt/.csvに記録、の流れで実装しようと思うが問題点は？

**回答：**
- 原因の見立てが違う可能性大：`Post EOF` はレート制限（クールタイム）ではなくGoのHTTP keep-alive接続の使い回しが原因（実装ポイント参照）。「何回まで送れるか測る」調査は空振りの公算。逆にリトライは新しい接続を開き直すのでこの原因にドンピシャで効く
- リトライ対象を区別：通信エラー（`err`）はリトライ／レスポンスの `Error`（重複など）は即失敗
- 戻り値の設計変更が必要：今の `addCard` は何も返さないので、ループ側で成否を判定するには `error` を返す形に
- 共通化のチャンス：同じPOST処理が6関数にコピペされている。`ankiInvoke` 1つにまとめてリトライをそこに書けば修正は1箇所で済む
- 失敗ファイルは `-f` で再投入できる形（1行1単語）にし、追記で。リトライ前に少し待つ（バックオフ）

### 2026-06-04
**質問：** 原因と思われるkeep-alive接続の使い回しについて詳しく教えて

**回答：**
- HTTP/1.1は接続を毎回開くと遅いので、一度開いたTCP接続を次のリクエストで使い回す（keep-alive）。Goの `http.Post` はデフォルトでこれをやる（接続プール）
- 接続を閉じる判断はサーバーとクライアントが別々に持つ。AnkiConnectが先に閉じた接続をGoが「まだ使える」と思って再利用 → 返事が来る前に切れていて `EOF`
- 単発実行では使い回す機会がなく起きない、連続実行で確率が上がる（タイミング依存の競合状態）。「10回試しても再現しない」のもこれで説明がつく
- 対策A：EOFが出たらリトライ（新接続が開かれ直る）／対策B：`DisableKeepAlives: true` で使い回し自体を止める。実用上は両方入れると堅牢

### 2026-06-04
**質問：** リトライってどう書く？（やさしめに）

**回答：**
- 「失敗したらもう一回やる」だけ。`for i := 0; i < 3` で最大3回、成功したら `break`/`return` で抜ける。抜けなければ次の回＝もう一回試す
- リトライ前に `time.Sleep(500 * time.Millisecond)` で少し待つと成功率が上がる
- 上限を決めて無限ループを防ぐ。3回ダメなら諦めて失敗記録へ
- 通信エラー（`err`）だけリトライ、Anki側の拒否（`ankiResp.Error`）はリトライせず即終了、と2種類の失敗を分けるのがコツ

### 2026-06-04
**質問：** 全関数にリトライを書くのは面倒。HTTPリクエストを送る関数を作ってそこで統一するってこと？（Invokeにまとめる意味）

**回答：**
- その通り。6関数で共通なのは「JSON化→POST→リトライ→body読む」、違うのは action名・params・レスポンスの型だけ
- 共通部分を `ankiInvoke(req any) ([]byte, error)` にまとめ、リトライはその中に1回だけ書く。各関数は `req` を作って渡し、返ってきた body を自分用の型に `Unmarshal` するだけ
- レスポンスの `Result` の型は関数ごとに違う（`[]string`/`[]int64`/`*int64`）ので、読み取りは各関数に残す（共通化しすぎるとジェネリクスの話になる）
- リトライ回数を変えたい時は `ankiInvoke` 1箇所だけ直せば全関数に反映（DRY原則。「レジ係1人」のイメージ）
- 進め方：`ankiInvoke` を作る→まず `addCard` 1つで動作確認→残りを1つずつ置き換え

### 2026-06-04
**質問：** `resp` をループ内で `:=` で作ったら、ループの外の `resp.Body.Close()` でエラーになる。なぜ？

**回答：**
- Goでは `{ }` の中で `:=` で作った変数はその中でしか使えない。`for` を抜けると `resp` は消えるので `undefined: resp` になる
- 直し方：`var resp *http.Response` をループの外で宣言し、成功時に `resp = r` で外の箱に入れる。さらに3回失敗時のため `if resp == nil { return }` を足す（nilのまま `resp.Body.Close()` すると今度こそpanic）

### 2026-06-04
**質問：** `fail` 関数の準備中。失敗単語をどこから取る？ 全体変数で持つのと各関数から取るのとどっちがいい？

**回答：**
- どちらでもなく `Run` のループ内ローカル `[]string` がベスト。グローバルは「いつ空にするか」問題・隠れた状態で避ける。各関数に持たせると役割が増えて複雑
- 「どの単語か」と「成否」を同時に知るのはループだけ。失敗時に `failedWords = append(...)` で集め、ループ後に1回だけファイルへ書く
- `fail` は `[]string` を受け取ってファイルに書くだけの関数にする

### 2026-06-04
**質問：** `addCard` をエラーを返すようにしたい。返り値は何にする？

**回答：**
- `error` 一つでOK。成功は `return nil`、失敗は `return エラー`
- `addCard` には失敗が2か所：① `ankiInvoke` の `err` → そのまま `return err` で横流し ② `ankiResp.Error != nil`（Anki側拒否）→ `fmt.Errorf("anki error: %s", *ankiResp.Error)` で包んで返す
- 渡したいデータは無く成否だけ知りたいので `error` 1つで十分。呼び出し側は `if err := addCard(...); err != nil { ... }` で受ける
- `isNote` は今 `bool` なので、エラーも返すなら `(bool, error)` 化を検討（addCardの後で）

### 2026-05-31
**質問：** CSV読み込みで `Error: failed to read CSV: record on line 2: wrong number of fields` が出た。意味は？

**回答：**
- 「2行目のセル数が他と違う」というエラー。`encoding/csv` は「全行が同じセル数のはず」という前提を持つ（表データ想定）
- `ReadAll()` は1行目のセル数を基準にし、以降の行が違うとエラーにする。テストCSVは行ごとにセル数がバラバラ（3,2,1）だったため2行目で発生
- 今回は「単語を並べるだけ」で行ごとにセル数が違うのが自然なので、列数チェックを無効化する
- 解決：`reader := csv.NewReader(file)` の直後に `reader.FieldsPerRecord = -1` を足す（`-1` = チェックしない）。これでバラバラな行数でも読める → 動作確認OK

### 2026-05-31
**質問：** ファイル取り込み実行中に `panic: Post "http://127.0.0.1:8765": EOF`（`isNote` の `http.Post` 失敗で panic）が出た。原因は？

**回答：**
- コードのバグではなく一時的な通信の途切れ（AnkiConnectがその1回だけ応答せず接続が切れた）。`EOF` = 返事が来る前に接続が閉じられた、の意味
- 10回ほど再試行したが再現しなかったため、原因は連続リクエストの一過性の不調と判断。対応は後回しにする
- ただし `isNote` 等が通信失敗で `panic` する作りなので、複数単語をループ処理中に1回失敗すると残り全部が巻き添えで止まる弱点がある（「後でやること」の panic→エラー表示対応がこれにあたる）。複数取り込みを仕上げた後に、panic をやめて `continue` でスキップする方向にする

### 2026-05-31
**質問：** `ankitango add [words] [deckName] [flags]` の意味は？

**回答：**
- cobra が自動生成する使い方の見本。`[words]`=単語/ファイル、`[deckName]`=デッキ名、`[flags]`=フラグを置ける印
- `[ ]` はプレースホルダ。`[words] [deckName]` は `add.go` の `Use` に書いた文字列がそのまま表示されている。`[flags]` は cobra が自動で足す
- `[flags]` が末尾なのは表示上の慣習で、値とセットなら前でも後ろでも置ける

### 2026-05-31
**質問：** `-f` を付けたら `flag needs an argument: 'f'` が出た。フラグは `add -f` みたいに付けるのでは？

**回答：**
- `-f`(`--file`)は string型フラグなので「`-f` の直後に値（ファイル名）」がセットで必要。`-f` を値なしで末尾に置いたのが原因
- 正しくは `add -f sample/words.txt MyDeck`
- bool型フラグ（`-h` など）は値不要で単独でOK。string型は値必須、という違い

### 2026-05-31
**質問：** args の数で分岐したい。3ならフラグあり/2なら単語/それ以外エラー、で行ける？

**回答：**
- フラグ（`-f 値`）は `args` に数えられない（cobra が解析して取り除く）。`add -f words.txt MyDeck` の `len(args)` は1
- なので「3ならフラグあり」では判定できない。フラグの有無は `filePath != ""` で見る
- 設計：A) `filePath != ""`→ファイルモード（deckName=args[0]、words=readWord）/ B) 空→args 2個必要（deckName=args[1]）、さらに args[0] の拡張子が .txt/.csv なら自動でファイル扱い、違えば単語1個
- ループ内の失敗は `return` でなく `continue`

### 2026-05-31
**質問：** addコマンドに.txtや.csvも取り込めるようにしたい。どうやる？

**回答：**
- 今の「1単語処理」を「ファイルなら中身を単語リストにして1個ずつ処理」に拡張するだけ
- `-f` フラグ方式と拡張子での自動判別方式の両方に対応する方針
- 標準ライブラリだけで可能：`bufio`（txt 1行ずつ）、`encoding/csv`（CSV解析）、`path/filepath`（拡張子判定）、`os`、`strings`
- 流れ：フラグ取得 → ファイルなら `readWord()` で `[]string` に → `for` で `isNote`→`generateWord`→`addCard`。失敗時は `return` でなく `continue`

### 2026-05-31
**質問：** 単語と「.txt/.csv」を自動判別する方式は曖昧と言ったけどどういうこと？

**回答：**
- 文字列の見た目（拡張子）だけで判断するので、まれに意図と食い違うという意味
- ただしこのアプリは語彙を登録する用途なので `.txt`/`.csv` で終わる単語はまず来ない＝実害はほぼない
- フラグ方式（`-f`）は合図が明確で確実。CLIの作法としても一般的

### 2026-05-31
**質問：** ファイル指定は普通 `-f` でやるもの？

**回答：**
- はい、一般的。CLIでは「主役の対象＝引数」「入力方法の切り替え＝フラグ」と分ける慣習
- 例：`kubectl apply -f file.yaml`、`git commit -F msg.txt`、`grep -f patterns.txt`
- 例外は `cat file.txt` のようにファイル自体が主役のツール
- ankitango は「単語登録」が主役でファイルは補助なので `-f` が筋が通る

### 2026-05-31
**質問：** `addCmd.Flags().StringP("file", "f", "", "...")` の意味は？

**回答：**
- addコマンドに文字列フラグ `--file`(短縮形 `-f`)を登録する
- 引数：長い名前 / 短縮形 / デフォルト値 / 説明文
- `StringP` の `P` は「短縮形あり」。`String` だと `--file` のみ
- 値は `cmd.Flags().GetString("file")` で取り出す

### 2026-05-31
**質問：** `filePath, _ := cmd.Flags().GetString("file")` は何をしている？

**回答：**
- 登録済みの `--file`(`-f`)フラグに渡された値を取り出して `filePath` に入れている
- `-f words.txt` なら `"words.txt"`、付けなければ `""`（デフォルト値）
- `_` は2つ目の戻り値（フラグ未登録時のエラー）を捨てている
- 後で `if filePath != ""` でファイル指定の有無を判定できる

### 2026-05-31
**質問：** `os.Open` した後 `file` には何が入っている？

**回答：**
- 中身そのものではなく「ファイルを操作するための取っ手（ハンドル、`*os.File`）」
- 開いた時点では中身は未読込。`bufio`/`csv` に渡して初めて読み出す
- 読むたびに「今ここまで読んだ」位置が進む。だから for で次々読める
- ポインタなのは、読み手が読み進めると元の `file` の位置も一緒に動く必要があるため

### 2026-05-29
**質問：** 新しいバージョンをリリースしたらもう一度インストールコマンドを打てばいい？

**回答：**
- はい。`install.sh` は引数なしだと毎回GitHubの最新リリース（latest）を自動取得する
- `/usr/local/bin/ankitango` を上書きするので古いバージョンを消す必要はない
- 自動更新はされないので、新バージョンを使うときは手動で再実行する
- バージョン指定したい場合は `bash install.sh v1.0.2` のように引数を渡す
- `go install` で入れた場合は `go install github.com/TobiTakuma/ankitango@latest`

### 2026-05-29
**質問：** もう一度インストールしてもかぶって変なことにならない？

**回答：**
- ならない。`install.sh` は一時フォルダにDLしてバイナリを上書き（`mv`）し、一時フォルダは削除する
- 古いバイナリが残ったり2個になったりしない（最初から上書きなので重複しない）
- `~/.config/ankitango/config.json`（APIキー・言語設定）には触れないので設定はそのまま残る

### 2026-05-29
**質問：** 技術スタックってどういう意味？

**回答：**
- そのソフトを作るために使っている技術の組み合わせのこと
- 言語・フレームワーク/ライブラリ・外部API・データ保存方法・配布の仕組みなどを含む
- 技術を層（レイヤー）として積み上げるイメージから「スタック（積み重ね）」と呼ぶ
- 例: LAMPスタック = Linux + Apache + MySQL + PHP

### 2026-05-29
**質問：** `add.go` で `config.go` の `loadConfig` を呼び出すにはどうすればいい？

**回答：**
- 同じ `package cmd` に属しているのでそのまま `loadConfig()` と書くだけ
- `import` も不要

### 2026-05-29
**質問：** エラー処理で `return` を書くと「not enough return values」と言われる

**回答：**
- `generateWord()` の戻り値が `map[string]string` なのでエラー時にも何か返す必要がある
- `return map[string]string{}` で空のmapを返す

### 2026-05-29
**質問：** 空のmapを返した時に2つエラーが出てしまう

**回答：**
- `generateWord()` が空のmapを返した後、そのまま `addCard()` が呼ばれているため
- `addCmd` の `Run` 内で `if len(fields) == 0 { return }` のチェックを追加する

### 2026-05-29
**質問：** 複数の意味がある単語はどう扱う？

**回答：**
- 今の実装はAIが毎回1つの訳を選んで生成する
- 同じ単語を何度も追加すれば自然と複数の意味がカードになる
- 1回の実行で複数カード生成する方法もあるが後回しでOK（後でやるリストに追加済み）

### 2026-05-29
**質問：** 新しいバージョンをリリースしたあと古いバージョンは自動更新される？

**回答：**
- 自動更新はされない。ユーザーが手動で更新する必要がある
- `go install github.com/TobiTakuma/ankitango@latest` を再実行するか、GitHubリリースページから新バイナリをダウンロード

### 2026-05-29
**質問：** GoReleaserのGitHub Personal Access Tokenとは何？

**回答：**
- GitHubが本人であることを確認するためのパスワードのようなもの
- GitHub CLIで `gh auth login` 済みであれば `GITHUB_TOKEN` の設定は不要
- `gh auth status` で確認できる

### 2026-05-29
**質問：** `.goreleaser.yaml` と `.github/workflows/release.yml` は別物？

**回答：**
- `.goreleaser.yaml` → GoReleaserの設定ファイル（どのOS向けにビルドするかなど）
- `.github/workflows/release.yml` → GitHub Actionsのワークフロー（タグをpushしたときに自動でGoReleaserを実行する設定）
- 両方必要

### 2026-05-29
**質問：** インストールスクリプトでJSONファイルも作られる？

**回答：**
- スクリプトはバイナリのインストールのみ
- `~/.config/ankitango/config.json` は `ankitango config apikey` を初めて実行したときに自動で作られる

### 2026-05-28
**質問：** JSONファイルはどこで作られるの？

**回答：**
- `os.WriteFile` がファイルを作る（なければ作成、あれば上書き）
- フォルダは自動で作られないので `os.MkdirAll` で先に作る必要がある

### 2026-05-28
**質問：** `MkdirAll` は毎回フォルダを作るの？

**回答：**
- フォルダが既にあっても何もしない（エラーにならない）
- 毎回呼んで大丈夫

### 2026-05-28
**質問：** `cfg` って何？

**回答：**
- `config` の略で慣習的によく使われる変数名
- `config` とそのまま書いても動くが、Goでは短い変数名が好まれる

### 2026-05-28
**質問：** `Config` って何？

**回答：**
- 設定データをまとめて持つための struct
- JSONファイルの構造をGoの struct として定義したもの
- `add.go` の `AnkiRequest` と同じ考え方

### 2026-05-28
**質問：** ファイルの読み書きはどうやるの？configファイルは勝手に作られるの？

**回答：**
- 読む: `os.ReadFile(path)`、書く: `os.WriteFile(path, data, 0644)`
- ファイルは `WriteFile` が自動で作る（なければ作成、あれば上書き）
- フォルダは自動で作られないので `os.MkdirAll` で先に作る必要がある
- `MkdirAll` はフォルダが既にあっても何もしない（エラーにならない）

### 2026-05-28
**質問：** `json.Unmarshal(data, &cfg)` の `&` の意味は？

**回答：**
- `&` はその変数のアドレス（メモリ上の場所）を渡す
- `Unmarshal` は結果を引数に直接書き込む設計なので、書き込み先を `&` で教える必要がある
- `&` がないと `cfg` は空のままになる
- `json.Marshal` は値を返す設計なので `&` 不要。`Unmarshal` だけ必要

### 2026-05-28
**質問：** `saveConfig` は差分だけ保存される？langを設定した後にAPIKeyを保存したらlangは消える？

**回答：**
- 差分保存ではなく毎回全体を上書きする
- 1つだけ更新したい場合は「loadConfig → 1つだけ書き換え → saveConfig」の流れにする
- こうすることで他のフィールドは消えずに残る

### 2026-05-28
**質問：** config.goはどんなふうに書くか？

**回答：**
- `~/.config/ankitango/config.json` にJSONで保存する方針
- `Config` structを定義して `loadConfig()` / `saveConfig()` で読み書きする
- `configCmd` のサブコマンドとして `apikey`, `lang`, `show` を定義する
- サブコマンドは `configCmd.AddCommand()` で追加し、`root.go` の `init()` で `configCmd` をrootに追加する
- ファイルが存在しない場合は空のConfigを返す（エラーにしない）

### 2026-05-28
**質問：** 変更ログの「2回目」「3回目」はどういう意味？

**回答：**
- 同じ日に複数のエントリがあったため、順番を区別するために付けた連番
- 意味がわかりにくいので、時刻を入れるか1日のエントリをまとめる方針に変更する

### 2026-05-28
**質問：** addとlistコマンドを同じファイルに書いているけど分けた方がいい？

**回答：**
- 今の規模なら分けなくて大丈夫
- コマンドが5個以上、または1ファイルが300〜400行を超えたら分けるタイミング

### 2026-05-28
**質問：** 配布する時のAPIキー設定はどうなる？

**回答：**
- 環境変数をユーザーが設定する方法（シンプル）とconfigコマンドで設定する方法がある
- 今は環境変数方式で十分。configコマンドはPhase 3で実装予定
- `godotenv` を外して `os.Getenv()` だけにすれば、どこからでも動く

### 2026-05-25
**質問：** ロジックは全部Goに移植して大丈夫？OpenAIのAPI以外も使えるようにしたいけど、AnkiConnectとかも使えるの？

**回答：**
- GoからOpenAI APIは公式SDKあり（`sashabaranov/go-openai`）、問題なく使える
- AnkiConnectはHTTP POSTでJSONをやり取りするだけなので、Goの標準ライブラリ（`net/http`）で完全に対応可能
- 他のLLM API（Claude、Geminiなど）も同様にHTTPリクエストで使える
- ロジックの全Go移植に技術的な問題はない

### 2026-05-25
**質問：** コマンドラインから直接操作するものとTUIの両方を作る予定。コマンドオプションはどれぐらい必要か？

**回答：**
- `anki add <word>` がメインコマンド
- オプションは `--deck`（デッキ指定）、`--lang`（カード方向 en-ja/ja-en）、`--dry-run`（プレビューのみ）の3つで十分
- サブコマンドは `list decks`、`list models`、`config`（デフォルト設定保存）、`tui`（TUI起動）
- `config` でデフォルト値を保存しておくと毎回オプションを打たなくて済む
- オプションは合計6〜7個程度で十分、使いながら追加する方針

### 2026-05-25
**質問：** AnkiConnectって言語は関係ない？JSONをいい感じに送ればいける？

**回答：**
- AnkiConnectはAnki内で動くローカルのHTTPサーバー（127.0.0.1:8765）
- 言語は関係なく、JSONをPOSTするだけで動く
- PythonでやっていたことをGoの `net/http` で書けばそのまま動く

### 2026-05-25
**質問：** PythonのrequestsライブラリのHTTP GETをGoで書き直すには何を使えばいい？

**回答：**
- `net/http`（標準ライブラリ）だけで完結
- `http.Get(url)` でGETリクエスト、`io.ReadAll(resp.Body)` でbody取得、`encoding/json` でパース
- 外部ライブラリ不要

### 2026-05-25
**質問：** `http.Get()` はどういう関数？何が返ってくる？

**回答：**
- `resp, err := http.Get(url)` の形で2つ返ってくる
- `resp` は `*http.Response`型。`resp.StatusCode`（ステータスコード）と `resp.Body`（レスポンスの中身、`io.Reader`型）を持つ
- `err` は通信自体の失敗（404などサーバーが返すエラーはここに来ない）
- `resp.Body` は使用後に `resp.Body.Close()` で閉じる必要がある

### 2026-05-25
**質問：** POSTって何？

**回答：**
- HTTPリクエストの種類の一つ
- GETは「データをくれ」、POSTは「データを送るから処理してくれ」
- AnkiConnectへのカード追加はPOSTを使う（命令とデータを一緒に送るから）
- ブラウザでURLを打つ＝GET、ログインボタンを押す＝POSTのイメージ

### 2026-05-25
**質問：** http.Postを使うには3つアイテムが必要というエラーが出ている。urlとjsonは分かるが3つ目は何？

**回答：**
- 3つ目はContent-Type（`"application/json"` 固定）
- `http.Post(url, "application/json", body)` の形で使う
- `body` は文字列をそのまま渡せないので `bytes.NewBuffer(jsonデータ)` で包む必要がある

### 2026-05-25
**質問：** GoでJSONはどう扱う？

**回答：**
- GoはJSONをそのまま書けない。`encoding/json` パッケージで変換する
- `json.Marshal(v)` でGoの値 → JSONのバイト列に変換
- mapを使う方法とstructを使う方法がある。ネストが深い場合はstructの方が読みやすい
- 変換したバイト列を `bytes.NewBuffer(jsonData)` で包んで `http.Post` に渡す

### 2026-05-27
**質問：** タグもfieldsに含められる？

**回答：**
- できない。AnkiConnectの仕様でfieldsとtagsは別のキーになっている
- OpenAIに `tags` も一緒に生成させて `generateWard()` の戻り値を `(map[string]string, []string)` の2つにするのが一番シンプル

### 2026-05-27
**質問：** `map[string]string` とは何？

**回答：**
- キーと値のペアを持つGoのデータ型
- `map[キーの型]値の型` という形で型を指定する
- 今回はキーも値も文字列なので `map[string]string`
- `fields["Front"]` のようにキーで値を取り出せる

### 2026-05-27
**質問：** generateWardで生成した結果をaddCardに渡すにはどうすればいい？

**回答：**
- `generateWard()` の戻り値を `map[string]string` にして4フィールドのmapを返す
- `main()` で `fields := generateWard(...)` として受け取り `addCard(fields)` に渡す
- `addCard()` も `func addCard(fields map[string]string)` の形に変える

### 2026-05-28
**質問：** コマンドの `-l` のような短縮形フラグはどう作る？

**回答：**
- `addCmd.Flags().StringP("deck", "d", "", "デッキ名")` のように `StringP` を使う
- `P` が「ショートバージョンあり」の意味
- `--deck 英単語` と `-d 英単語` が同じ意味になる
- `init()` 内で定義し、`Run` 内で `cmd.Flags().GetString("deck")` で取得する

### 2026-05-28
**質問：** GoのforループはどうやってかくのKa？

**回答：**
- `for i := 0; i < 3; i++` → 回数指定
- `for 条件 {}` → whileのような使い方
- `for {}` → 無限ループ（breakかreturnで抜ける）
- `for i, v := range slice {}` → スライス・mapの繰り返し

### 2026-05-27
**質問：** cobraのコマンド定義 `var addCmd = &cobra.Command{...}` はどういう書き方？

**回答：**
- cobraは「コマンドの設定をstructで書く」設計
- `cobra.Command` というstructに名前・説明・実行処理をまとめて書く
- `Run` フィールドに「コマンドが呼ばれたときに実行する関数」を渡す
- `anki add apple` と打つとcobraが `Run` の中を実行してくれる
- 今まで `func main()` に書いていた処理をcobraに登録する形に変わるだけ

### 2026-05-27
**質問：** IsModelとaddNewModelの実行タイミングはどうするか？

**回答：**
- `IsModel()` は毎回起動時に走らせる（途中でモデルを消した場合にも対応）
- `addNewModel()` は `IsModel()` が false のときだけ実行する
- `if !IsModel() { addNewModel() }` の形で main に入れる

### 2026-05-27
**質問：** なぜ `isCloze` という名前なのか？

**回答：**
- AnkiConnectのAPIがそのキー名を `isCloze` と決めているため変えられない
- Clozeとはが穴埋め問題形式のカードのこと（例: "The capital of France is [...]"）
- `isCloze: false` は「通常の表裏カード」という意味

### 2026-05-27
**質問：** PythonのcreateModelでどこでHTTPを送っているかわからない

**回答：**
- `invoke()` という共通関数の中でHTTPを送っている（13行目）
- `create_model()` は `invoke()` を呼ぶだけなのでHTTPが見えない
- Goでは今の他の関数と同じように関数の中に直接HTTP処理を書けばOK

### 2026-05-26
**質問：** AnkiのフィールドをAnkiRequestから外部で作ることはできる？表・英語例文・裏・日本語例文の4フィールドのカードを作りたい。

**回答：**
- AnkiConnectの `createModel` アクションで外部からカードタイプ（モデル）を作れる
- `python/test.py` にすでに同じコードがある
- `createModel` で一度モデルを作っておけば、あとは `addNote` でそのモデルを指定するだけ
- 進め方: createModel関数を作る → プロンプトを4フィールド対応に整える → addNoteを修正

### 2026-05-26
**質問：** req.Headerとは何？

**回答：**
- HTTPリクエストはヘッダー（付箋）とボディ（データ本体）の2つで構成される
- `Content-Type: application/json` → 送るデータがJSON形式だとサーバーに伝える
- `Authorization: Bearer <key>` → APIキーで身分証明する
- 荷物（ボディ）に貼る伝票（ヘッダー）のイメージ

### 2026-05-26
**質問：** GoでOpenAI APIを使う方法を詳しく教えて。普通はどうやるか、今回はどうやるかも含めて。

**回答：**
- 普通はSDK（`sashabaranov/go-openai`）を使うが、今回は複数API対応のため生HTTPで統一
- AnkiConnectと違う点は2つ：`http.NewRequest` を使う（ヘッダー追加のため）、APIキーをAuthorizationヘッダーにセットする
- APIキーは `os.Getenv("OPENAI_API_KEY")` で取得
- 流れはAnkiConnectと同じ（JSON作成→リクエスト作成→ヘッダーセット→送信→レスポンスパース）

### 2026-05-26
**質問：** APIや各AIの方式についてもう少し教えて

**回答：**
- APIはサービスの「受付窓口」。決められた形式でJSONを送ると結果が返ってくる
- OpenAI・Gemini・Claude どれもHTTPリクエストでJSONを送受信する点は同じ
- 違いはURL・JSONの構造・キー名だけ
- AnkiConnectにPOSTしたのと本質的にやることは同じ

### 2026-05-26
**質問：** SDKと生HTTPの違いは？

**回答：**
- 生HTTP: 自分でJSONを組み立てて送る方法。AnkiConnectでやったのと同じ
- SDK: API提供会社が用意したライブラリ。細かい部分を隠してシンプルに書ける
- 複数APIに対応する今回のケースでは、生HTTPで統一する方が一貫性があってシンプル

### 2026-05-26
**質問：** OpenAI以外にGeminiやClaudeにも対応させたい。どうすればいい？

**回答：**
- Goのinterfaceを使って共通の窓口（`LLMClient`）を定義し、各APIをそれぞれ実装する方針が良い
- OpenAI・GeminiはSDKあり、ClaudeはGoの公式SDKなしなので生HTTP
- 全部生HTTPで統一する方法もある（AnkiConnectと同じ要領）
- まずどのAPIから実装するか・SDKを使うか生HTTPで統一するかを決める

### 2026-05-26
**質問：** paramsって何？

**回答：**
- AnkiConnectのAPIが決めているJSONのキー名（parametersの略）
- 「どのデッキに、どのモデルで、どんな内容を追加するか」を入れる場所
- 自分で決めた名前ではなくAnkiConnectの仕様に合わせている
