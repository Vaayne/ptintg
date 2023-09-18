package bot

import (
	"fmt"
	"log/slog"
	"ptintg/pkg/pt/mteam"
	"strings"
	"text/template"

	tele "gopkg.in/telebot.v3"
)

const ItemTmp = `
Title: {{.Name}}

大小: {{.Size}}
下载数: {{.Leechers}}
做种数: {{.Seeders}}
发布时间: {{.PublishedOn}}
下载: /mteam{{.ID}}
`

func OnMteamSearch(c tele.Context) error {

	query := c.Data()
	if query == "" {
		slog.Warn("empty keyword")
		return c.Send("Invalid search keyword, please use `/search {query}`, for example `/search 肖申克的救赎`")
	}

	items, err := mtClient.Read(&mteam.Option{
		Search: query,
	})
	if err != nil {
		msg := fmt.Sprintf("search from mteam error: %s", err)
		slog.Error(msg)
		return c.Send(msg)
	}

	if len(items) == 0 {
		msg := fmt.Sprintf("No result for %s", query)
		slog.Warn(msg)
		return c.Send(msg)
	}

	for _, item := range items {
		buf := &strings.Builder{}
		t := template.Must(template.New("mteam item").Parse(ItemTmp))
		if err := t.Execute(buf, item); err != nil {
			slog.Error("write data to template error", "data", item)
			continue
		}
		if err := c.Send(buf.String()); err != nil {
			slog.Error("send data to user error", "err", err)
		}
	}

	return nil
}

func OnMTeamDownload(c tele.Context, text string) error {
	data := text[6:]
	if len(data) < 3 {
		return c.Reply(fmt.Sprintf("Invalid mteam item %s", data))
	}
	slog.Info("download mteam item", "item_id", data)
	downlaodUrl := mtClient.BuildDownlaodUrl(data)
	filename, err := mtClient.Download(downlaodUrl)
	if err != nil {
		slog.Error("download torrent file error", "url", downlaodUrl, "err", err)
		return c.Reply("download torrent file error")
	}

	if err := qbCLient.AddTorrentFromFile(filename, nil); err != nil {
		slog.Error("download torrent from file error", "filename", filename, "err", err, "url", downlaodUrl)
		return c.Reply("download torrent error")
	}

	return c.Reply("success downlaod " + mtClient.BuildDetailUrl(data))
}
