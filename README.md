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
3. An API key for OpenAI or Gemini.   
(In my case, I tried more than 300 words, but it only cost $0.02. If you're concerned about API costs, I recommend the Gemini API. Gemini has a generous free API tier (verify your age and obtain an API key).)

      | provider |where to get an API key                       |
      | -------- |--------------------------------------------- |
      | openai   |https://platform.openai.com/api-keys          |
      | gemini   |https://aistudio.google.com/apikey            |
   

## installation and setup

### 1) Install ankitango in your computer
**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.sh | sh
```

If you are asked for a password, please enter the password for your device.  
If /usr/local/bin does not exist, you will need to create it.
```bash
sudo mkdir /usr/local/bin

# or

mkdir /usr/local
cd /usr/local
mkdir bin
```


**Windows (x86_64)**

```powershell
powershell -Command "Invoke-WebRequest -Uri https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.ps1 -OutFile install.ps1; .\install.ps1"
```


If you are asked for a password, please enter the password for your device.

### 2) Set up
Before you begin using ankitango, run three commands.  
(Detailed instructions are provided below)

```bash
# 1. Set the LLM provider and API key (model is optional)
ankitango config apikey <provider> <key> [model] 
# 2. Set language
ankitango config lang <fromLang> <toLang>      
# 3. Make sure your settings are reflected
ankitango config show                      
```

#### Supported LLMs

ankitango talks to each provider through an OpenAI-compatible endpoint, so the list of providers can grow over time.

| provider | default model      |
| -------- | ------------------ |
| openai   | `gpt-4o-mini`      |
| gemini   | `gemini-2.5-flash` |

Pick one when you set your key:

```bash
ankitango config apikey <provider> <key>
```

If the default model is outdated or you want a different one, pass it as the third argument (no need to update the tool):

```bash
ankitango config apikey gemini "AIza...xyz" gemini-2.5-pro
```
#### Config Command Example Collection 

```
# Supported providers: openai, gemini
# The endpoint and a default model are set automatically per provider.

# ex) OpenAI
ankitango config apikey openai "sk-...jGcA"      # when you use it, paste the whole api key

# ex) Gemini
ankitango config apikey gemini "AIza...xyz"

# ex) specify a model explicitly (e.g. when the default model is outdated)
ankitango config apikey gemini "AIza...xyz" gemini-2.5-flash

# ex) Language Selection
ankitango config lang English Japanese

# ex) result of "ankitango config show"
LLM Model: gemini-2.5-flash
APIkey: AQ....Tlpg
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
1. if you want add word.  
Ex) You want add "word" to "wordDeck" deck.
   ```bash
   ankitango add word wordDeck
   ```
   If successful, you get results like this:
   ``` bash
   Generating...
   {
      "Front": "word",
      "Front_Sentence": "Please tell me a word.",
      "Back": "言葉",
      "Back_Sentence": "一言教えてください。"
   }
   Success!
   ```
2. if you want add idiom or chose a deck that include space. You need double quote.
   ```bash
   ankitango add "hello world" "word deck"
   ```

3. add many words at once from a `.txt` or `.csv` file with the `-f` flag.
   ```bash
   ankitango add -f words.txt wordDeck   # .txt : one word per line
   ankitango add -f words.csv wordDeck   # .csv : one word per cell (number of columns can vary)
   ```
   In file mode you only pass the deck name. Words that fail (e.g. the connection to Anki drops) are saved to `fail.txt` in the same folder as the input file, so you can retry them later with `-f`.
### config "subcommand"
1. setting provider and api key (model is optional)
   ```bash
   ankitango config apikey openai "sk-...Gkacv"
   ankitango config apikey gemini "AIza...xyz"
   ```
   Supported providers: `openai`, `gemini`. The endpoint and a default model are chosen automatically.
   If the default model is outdated or unavailable, pass a model name as the third argument:
   ```bash
   ankitango config apikey gemini "AIza...xyz" gemini-2.5-flash
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