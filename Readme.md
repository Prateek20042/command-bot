# Go Command Bot

A Go-based CLI tool that analyzes work commands using LLM (Ollama) to extract structured instructions, actions, and entity resolutions.

## Features

- 🧠 LLM-powered command analysis
- ✂️ Extracts clean instructions and actions
- 🏢 Resolves departments from context clues
- 📝 Maintains execution logs
- 🛠️ Easy to extend with new inference rules

## Prerequisites

- Go 1.18+
- [Ollama](https://ollama.ai/) running locally
- (Optional) Git for cloning

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/command-bot.git
cd command-bot
```
2. Install dependencies:
   ```bash
   go mod tidy
   ollama serve
   ```
3. Run the bot:
   ```bash
   go run main.go
   ```
