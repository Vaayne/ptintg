package bot

import (
	"log/slog"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func OnText(c tele.Context) error {
	text := c.Text()
	slog.Info("data", "text", text)

	if strings.HasPrefix(text, "/mteam") {
		return OnMTeamDownload(c, text)
	}

	return c.Reply(text)
}
