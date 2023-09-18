package bot

import (
	"log/slog"
	"net/http"
	"ptintg/internal/pkg/config"
	"ptintg/pkg/pt/mteam"

	"github.com/Vaayne/aienvoy/pkg/cookiecloud"
	"github.com/autobrr/go-qbittorrent"
	tele "gopkg.in/telebot.v3"
)

var mtClient *mteam.MTeam
var qbCLient *qbittorrent.Client

func init() {
	initQBittorrent()
	initMTeam()
}

func initQBittorrent() {
	qbConfig := config.GetConfig().QBittorrent
	qbCLient = qbittorrent.NewClient(qbittorrent.Config{
		Host:     qbConfig.Host,
		Username: qbConfig.User,
		Password: qbConfig.Pass,
	})
	if err := qbCLient.Login(); err != nil {
		slog.Error("login to qBittorrent error", "err", err)
		panic("Login to qbittorrent error")
	}
}

func getCookies(domain string) []*http.Cookie {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)
	getCookie := func(domain string) []*http.Cookie {
		cookies := make([]*http.Cookie, 0)
		cks, _ := cc.GetCookies(domain)
		for _, ck := range cks {
			cookies = append(cookies, &http.Cookie{
				Name:  ck.Name,
				Value: ck.Value,
			})
		}
		return cookies
	}
	return getCookie(domain)
}

func initMTeam() {
	mtCookies := getCookies(mteam.DOMAIN)
	slog.Debug("get cookie for "+mteam.DOMAIN, "cookies", mtCookies)
	mtClient = mteam.New(mtCookies...)
}

func registerHandlers(b *tele.Bot) {
	b.Handle("/search", OnMteamSearch)
	b.Handle(tele.OnText, OnText)
}

func Serve() {
	b, err := tele.NewBot(tele.Settings{
		Token: config.GetConfig().Telegram.Token,
	})
	if err != nil {
		slog.Error("init telegram bot error", "err", err)
		return
	}

	registerHandlers(b)
	slog.Info("start telegram bot")
	b.Start()
}
