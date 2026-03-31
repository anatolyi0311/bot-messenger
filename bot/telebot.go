package bot

import (
	"anatolyi0311/bot-messenger/internal/config"
	"time"

	tb "gopkg.in/telebot.v3"
)

func Bot(cfg *config.Config) {
	b, err := tb.NewBot(tb.Settings{
		Token:  cfg.Bot.Token,
		Poller: &tb.LongPoller{Timeout: 30 * time.Second},
	})
	if err != nil {
		panic(err)
	}
	b.Handle("/start", func(c tb.Context) error {
		return c.Send("Hello world!")
	})
	b.Start()
}
