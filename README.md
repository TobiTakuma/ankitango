# ankitango

CLIツールで単語を入力するだけで、翻訳と例文を自動生成してAnkiに追加できます。

## 必要なもの

- [Anki](https://apps.ankiweb.net/)
- [AnkiConnect](https://ankiweb.net/shared/info/2055492159)（Ankiのアドオン）
- OpenAI APIキー

## セットアップ

```bash
git clone https://github.com/TobiTakuma/ankitango.git
cd ankitango
```

`.env` ファイルをルートに作成：
```
OPENAI_API_KEY=your-api-key
```

## 使い方

Ankiを起動した状態で実行してください。

```bash
# 単語をAnkiに追加
go run main.go add <word> <deckName>

# 例
go run main.go add apple MyDeck
go run main.go add "look up" MyDeck  # スペースを含む場合はクォートで囲む

# デッキ一覧を表示
go run main.go list
```

## カードの形式

| フィールド | 内容 |
|-----------|------|
| Front | 英語の単語 |
| Front_Sentence | 英語の例文 |
| Back | 日本語の訳 |
| Back_Sentence | 日本語の例文 |

初回実行時に `AddAnkiCLI` というカードタイプが自動で作成されます。
