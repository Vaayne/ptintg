package config

import "github.com/Vaayne/aienvoy/pkg/config"

var globalConfig = &Config{}

func init() {
	config.Load(globalConfig)
}

func GetConfig() *Config {
	return globalConfig
}

type Config struct {
	Telegram struct {
		Token string `yaml:"token"`
	}

	CookieCloud CookieCloud
	QBittorrent QBittorrent
}

type CookieCloud struct {
	Host string
	UUID string
	Pass string
}

type QBittorrent struct {
	Host string
	User string
	Pass string
}
