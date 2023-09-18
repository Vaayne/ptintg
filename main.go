package main

import (
	"os"
	"ptintg/internal/bot"

	"log/slog"

	"github.com/Vaayne/aienvoy/pkg/loghandler"
)

func init() {
	initLog()

}

func initLog() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})
	slog.SetDefault(slog.New(loghandler.NewHandler(handler)))
}

func main() {

	bot.Serve()

}
