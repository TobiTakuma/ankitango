# ankitango

A CLI tool that automatically generates translations and example sentences using AI, then adds them to Anki.

## Requirements

- [Anki](https://apps.ankiweb.net/)
- [AnkiConnect](https://ankiweb.net/shared/info/2055492159) (Anki add-on)
- OpenAI API key

## Setup

**Using go install**
**macOS / Linux**
```bash
curl -fsSL https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.sh | sh
```

**Windows (x86_64)**
```powershell
powershell -Command "Invoke-WebRequest -Uri https://raw.githubusercontent.com/TobiTakuma/ankitango/main/install.ps1 -OutFile install.ps1; .\install.ps1"
```

Configure your settings:
```bash
ankitango config apikey <key>              # Set OpenAI API key
ankitango config lang <fromLang> <toLang>  # Set language (e.g. English Japanese)
ankitango config show                      # Show current settings

ex)
ankitango config apikey "sh...jGcA" # when you use it, paste all api key
ankitango config lang English Japanese 
```

## Usage

Make sure Anki is running before executing commands.

```bash

# List all decks
ankitango list

# Add a word to Anki
ankitango add <word> <deckName>

# Examples
ankitango add apple MyDeck
ankitango add "look up" MyDeck   # Use quotes for words with spaces
ankitango add "apple" "word list"

```

## Card Format

| Field | Content |
|-------|---------|
| Front | Word |
| Front_Sentence | Example sentence |
| Back | Translation |
| Back_Sentence | Translated example sentence |

A card type named `ankitango` is automatically created on first run.
