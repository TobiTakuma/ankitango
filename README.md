# ankitango
A CLI tool that automatically generates translations and example sentences using AI, then adds them to Anki.

```bash
ankitango add "choice" "test"

Generating...
{
  "Front": "choice",
  "Front_Sentence": "It's important to make the right choice when it comes to your career.",
  "Back": "選択",
  "Back_Sentence": "キャリアに関して正しい選択をすることが重要です。"
}
Success!
```
## Requirements 

1. [Anki](https://apps.ankiweb.net/)
2. [AnkiConnect](https://ankiweb.net/shared/info/2055492159) (Anki add-on)
   
   tools -> Add-ons -> Get Add-ons... -> type 2055492159 -> restart app
3. OpenAI API key
   

## Setup/installation command

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.sh | sh
```

If you are asked for a password, please enter the password for your device.


**Windows (x86_64)**

```powershell
powershell -Command "Invoke-WebRequest -Uri https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.ps1 -OutFile install.ps1; .\install.ps1"
```


If you are asked for a password, please enter the password for your device.

  

Configure your settings:

```bash
ankitango config apikey <key> # Set OpenAI API key
ankitango config lang <fromLang> <toLang> # Set language (e.g. English Japanese)
ankitango config show # Show current settings

# ex1)
ankitango config apikey "sh...jGcA" # when you use it, paste all api key
ankitango config lang English Japanese

# ex of "ankitango config show") 
APIkey: sk-...adoV
Lang: "English" to "Japanese"
```

  
## Usage

Make sure Anki is running before executing commands.

| command | discription                                               |
| ------- | --------------------------------------------------------- |
| add     | Generate words and example sentences and add them to Anki |
| config  | Manage the config data. It has 3 subcommand.              |
| list    | List all deck in Anki                                     |
| help    | help command                                              |
### add "word" "deckName"
1. if you want add word.Ex) You want add "word" to "wordDeck" deck.
   ```bash
   ankitango add word wordDeck
   ```

2. if you want add idiom or chose a deck that include space. You need double quote.
   ```bash
   ankitango add "hello world" "word deck"
   ```
### config "subcommand"
1. setting api key
   ```bash
   ankitango config apikey "sk-...Gkacv"
   ```
2. setting language
   ```bash
   ankitango config lang English Japanese
   ```
3. see your current setting
   ```bash
   ankitango config show
   ```
### list
1. list all your Anki deck
   ```bash
   ankitango list
   ```


## Card Format

| Field          | Content                     |
| -------------- | --------------------------- |
| Front          | Word                        |
| Front_Sentence | Example Sentence            |
| Back           | Translation                 |
| Back_Sentence  | Translated Example Sentence |

A card type named `ankitango` is automatically created on first run.