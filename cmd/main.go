package main

import (
	"anatolyi0311/bot-messenger/bot"
	"anatolyi0311/bot-messenger/internal/config"
)

func main() {
	cfg := config.LoadConfig()
	bot.Bot(cfg)
}
