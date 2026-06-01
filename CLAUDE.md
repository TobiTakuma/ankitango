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

- `panic(err)` をエラーメッセージ表示に変更（複数取り込みのループ中に1回の通信失敗で全体が死ぬのを防ぐ。`continue` でスキップする方向）
- 発音記号・音声データフィールドの追加（`Pronunciation`, `Audio` フィールドは存在するがAIが生成していない。生成内容に含める）
- 複数の意味・訳語への対応（1回の実行で複数カードを生成するか検討）
- 拡張子での自動判別（`-f` なしで `add words.txt MyDeck` と打ったら、args[0] が `.txt`/`.csv` なら自動でファイル扱いにする）：現状は `-f` フラグ必須。優先度低
- 他LLM対応（Claude / Gemini）：`LLMClient` interface で共通化する案あり

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

---

## 実装のポイント（学んだこと）

- AnkiConnectはHTTP POSTでJSONを送るだけ。言語は関係ない
- GoのJSON扱い: `json.Marshal()` でGoの値→JSONバイト列、`json.Unmarshal()` で逆変換
- `http.Post(url, "application/json", bytes.NewBuffer(jsonData))` の形で送る
- `resp.Body` は必ず `defer resp.Body.Close()` で閉じる
- AnkiConnectのレスポンスは `{"result": ..., "error": ...}` の形

---

## 質問ログ

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
