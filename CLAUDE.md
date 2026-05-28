# Anki自動追加アプリ CLAUDE.md

## 完成までの計画

| フェーズ | 内容 | 状態 |
|---------|------|------|
| 1. AnkiConnect基本機能 | `checkAnkiRunning()`, `getDeckName()`, `addNote()` | ✅ 完了 |
| 2. OpenAI API連携 | 単語を入力 → 翻訳・例文を生成 | ✅ 完了 |
| 3. 機能追加 | デッキ・モデル選択、重複チェック、カード方向管理、タグ自動付与 | 🔄 進行中 |
| 4. CLIコマンド化 | cobra導入、`anki add <word>` などのコマンド実装 | ✅ 完了 |
| 5. TUI | bubbletea導入、画面フロー実装（AIに任せる） | 未着手 |
| 6. 配布 | GoReleaserでリリース | 未着手 |

---

## CLIコマンド化の手順（cobra）

### 全体像
```
main.go        ← エントリーポイント。cmd.Execute() を呼ぶだけ
cmd/
  root.go      ← ルートコマンド（anki）の定義
  add.go       ← add サブコマンドの定義。ここから generateWard → addCard を呼ぶ
```

今 `main.go` にある関数（`generateWard`, `addCard` など）はそのままで、`main()` の中身だけ `cmd/add.go` に移す。

### 手順
1. `go get github.com/spf13/cobra` でインストール
2. `cmd/` フォルダを作成
3. `cmd/root.go` を作成（ルートコマンドの定義）
4. `cmd/add.go` を作成（add サブコマンド。今の main() の処理をここに移す）
5. `main.go` を `cmd.Execute()` だけに書き直す

### 完成後の使い方
```bash
anki add apple                    # 英→日でカード追加
anki add apple --deck 英単語      # デッキ指定（後で実装）
anki add apple --lang ja-en       # カード方向指定（後で実装）
```

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

### 2026-05-28 (4回目)
**質問：** addとlistコマンドを同じファイルに書いているけど分けた方がいい？

**回答：**
- 今の規模なら分けなくて大丈夫
- コマンドが5個以上、または1ファイルが300〜400行を超えたら分けるタイミング

### 2026-05-28 (3回目)
**質問：** 配布する時のAPIキー設定はどうなる？

**回答：**
- 環境変数をユーザーが設定する方法（シンプル）とconfigコマンドで設定する方法がある
- 今は環境変数方式で十分。configコマンドはPhase 3で実装予定
- `godotenv` を外して `os.Getenv()` だけにすれば、どこからでも動く

### 2026-05-28 (2回目)
**ユーザーの指示により:**
- `go install .` でローカルにバイナリをインストール
- `~/.zshrc` に `export PATH=$PATH:$HOME/go/bin` を追加してPATHを通した
- `ankitango` コマンドで直接起動できるようになった

### 2026-05-28
**ユーザー:**
- `list` コマンド実装：`ankitango list` でデッキ一覧を表示
- `isDeck()` 関数追加：存在しないデッキ名を指定した時にエラーメッセージを表示
- `getDeckName()` を `[]string` を返す関数に変更
- `addCard()` のレスポンス型を `*int64` に修正、成功メッセージを追加
- `.venv`, `.idea`, `python/` を整理・削除

**Claude:**
- 不要ファイル（`.venv`, `.idea`）を `python/` フォルダに移動
- README.md を書き直し（Goプロジェクト用に更新）

**ユーザーの指示により:**
- git push時にAPIキー漏洩エラー（`test.py` に古いキーが残っていた）
- `.git` を削除して `git init` からやり直す方針に決定

---

## 実装のポイント（学んだこと）

- AnkiConnectはHTTP POSTでJSONを送るだけ。言語は関係ない
- GoのJSON扱い: `json.Marshal()` でGoの値→JSONバイト列、`json.Unmarshal()` で逆変換
- `http.Post(url, "application/json", bytes.NewBuffer(jsonData))` の形で送る
- `resp.Body` は必ず `defer resp.Body.Close()` で閉じる
- AnkiConnectのレスポンスは `{"result": ..., "error": ...}` の形

---

## 質問ログ

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
