package mteam

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const DOMAIN = "kp.m-team.cc"

type MTeam struct {
	Domain  string
	cookies []*http.Cookie
}

func New(cookies ...*http.Cookie) *MTeam {
	return &MTeam{
		Domain:  DOMAIN,
		cookies: cookies,
	}
}
func (m *MTeam) BuildDownlaodUrl(id string) string {
	return fmt.Sprintf("https://%s/download.php?id=%s&https=1", m.Domain, id)
}

func (m *MTeam) BuildDetailUrl(id string) string {
	return fmt.Sprintf("https://%s/detail.php?id=%s&https=1", m.Domain, id)
}

func (m *MTeam) fetch(url string) (io.ReadCloser, error) {
	req, _ := http.NewRequest("GET", url, nil)

	// set cookies
	for _, cookie := range m.cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	// fetch data
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("fetch M-Team error", "err", err)
		return nil, err
	}

	slog.Debug("fetch M-Team", "url", url, "status", resp.StatusCode)
	return resp.Body, nil
}

func (m *MTeam) Download(url string) (string, error) {
	r, err := m.fetch(url)
	if err != nil {
		return "", fmt.Errorf("download from %s error: %w", url, err)
	}

	f, err := os.CreateTemp("", "mteam-*")
	if err != nil {
		return "", fmt.Errorf("create temp file error: %w", err)
	}

	_, err = f.ReadFrom(r)
	if err != nil {
		return "", fmt.Errorf("write torrent to file error: %w", err)
	}
	return f.Name(), nil
}

type Option struct {
	// common options
	IncludeDead    int `json:"incldead"`       // 0: 包括断种, 1: 仅包括活种， 2: 仅包括断种
	SpState        int `json:"spstate"`        // 种子促销相关: 0: 全部， 1: 普通， 2: 免费， 3: 2X, 4: 2X free, 5: 50%, 6: 2x 50%, 7: 30%
	InclBookmarked int `json:"inclbookmarked"` // 0: 全部, 1: 仅收藏, 2: 仅非收藏
	Page           int `json:"page"`           // 页数

	// options for search
	SearchMode int    `json:"searchmode"` // 0: AND, 1: OR， 2: 准确搜索
	SearchArea int    `json:"searcharea"` // 0, 1, 2: 标题搜索, 3: 发布者, 4: IMDB
	Search     string `json:"search"`     // 搜索关键字
}

func (o Option) toUrlValues() url.Values {
	vals := url.Values{
		"incldead":       {strconv.Itoa(o.IncludeDead)},
		"spstate":        {strconv.Itoa(o.SpState)},
		"inclbookmarked": {strconv.Itoa(o.InclBookmarked)},
		"page":           {strconv.Itoa(o.Page)},
	}
	if o.Search != "" {
		vals.Add("search", o.Search)
		vals.Add("searchmode", strconv.Itoa(o.SearchMode))
		vals.Add("searcharea", strconv.Itoa(o.SearchArea))
	}
	return vals
}

func NewOption() *Option {
	return &Option{
		IncludeDead: 1,
	}
}

func (m *MTeam) Read(opt *Option) ([]*Result, error) {
	if opt == nil {
		opt = NewOption()
	}
	u := url.URL{
		Scheme:   "https",
		Host:     m.Domain,
		Path:     "/torrents.php",
		RawQuery: opt.toUrlValues().Encode(),
	}

	r, err := m.fetch(u.String())
	if err != nil {
		return nil, err
	}
	defer r.Close()
	items := parseResult(r)
	return items, nil
}

type Result struct {
	ID          string // 种子 ID
	Category    string // 种子类型
	Title       string // 标题
	Name        string // 种子名称
	Comments    int    // 评论数
	PublishedOn string // 发布时间
	Size        string // 大小
	Seeders     int    // 做种人数
	Leechers    int    // 下载人数
	Snatched    int    // 完成人数
	Publisher   string // 发布者
}

func (r Result) DownloadUrl() string {
	return fmt.Sprintf("https://%s/download.php?id=%s&https=1", DOMAIN, r.ID)
}

func (r Result) DetailUrl() string {
	return fmt.Sprintf("https://%s/detail.php?id=%s&https=1", DOMAIN, r.ID)
}

func parseResult(r io.Reader) []*Result {
	items := make([]*Result, 0)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		slog.Error("new document error", "err", err)
		return items
	}

	doc.Find(".torrents tbody tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		findInt := func(td *goquery.Selection, name string) int {
			text := td.Find("a").First().Text()
			if text == "" {
				return 0
			}
			ret, err := strconv.Atoi(text)
			if err != nil {
				slog.Warn("parse int error", "name", name, "err", err, "text", td.Text())
				return 0
			}
			return ret
		}

		res := &Result{}
		s.Children().Each(func(i int, td *goquery.Selection) {
			switch i {
			case 0: // Type
				s := td.Find("a").First()
				t, ok := s.Attr("href")
				if ok {
					res.Category = strings.Split(t, "=")[1]
				}
			case 1: // ID, Title, Name
				td = td.Find(".embedded")
				res.Name = td.Text()
				res.Title = td.Find("b").First().Text()

				info := td.Find("a").First()
				r := regexp.MustCompile(`id=(\d+)`)
				match := r.FindStringSubmatch(info.AttrOr("href", ""))
				if len(match) > 1 {
					res.ID = match[1]
				}
				if res.ID == "" {
					return
				}
			case 2: // Comments
				res.Comments = findInt(td, "comments")
			case 3: // published datetime
				res.PublishedOn = td.Find("span").First().AttrOr("title", "")
			case 4: // size
				res.Size = td.Text()
			case 5: // Seeders
				res.Seeders = findInt(td, "seeders")
			case 6: // Leechers
				res.Leechers = findInt(td, "leechers")
			case 7: // Snatched
				res.Snatched = findInt(td, "snatched")
			case 9: // Publisher
				res.Publisher = td.Text()
			}
		})
		if res.ID != "" {
			items = append(items, res)
		}
		// slog.Debug("item detail", "item", res, "text", s.Text())
	})
	return items
}
