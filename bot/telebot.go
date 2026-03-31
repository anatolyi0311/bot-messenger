package bot

import (
	"time"

	tb "gopkg.in/telebot.v3"
)

func Bot() {
	b, err := tb.NewBot(tb.Settings{
		Token:  "YOUR_TELEGRAM_BOT_TOKEN",
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
