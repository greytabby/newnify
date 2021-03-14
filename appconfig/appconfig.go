package appconfig

import (
	"os"

	_ "github.com/GoogleCloudPlatform/berglas/pkg/auto"
)

type app struct {
	AllowOrigin string
	Port        string
}
type twitter struct {
	ConsumerAPIKey       string
	ConsumerAPIKeySecret string
	AccessToken          string
	AccessTokenSecret    string
}

var App = app{
	AllowOrigin: "http://localhost:3000",
	Port:        "7777",
}

var Twitter = twitter{
	ConsumerAPIKey:       "",
	ConsumerAPIKeySecret: "",
	AccessToken:          "",
	AccessTokenSecret:    "",
}

func getEnvOrDefault(name string, defaultValue string) string {
	envValue := os.Getenv(name)
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

func NewConfig() {
	App.AllowOrigin = getEnvOrDefault("ALLOW_ORIGIN", "http://localhost:3000")
	App.Port = getEnvOrDefault("PORT", "7777")
	Twitter.ConsumerAPIKey = getEnvOrDefault("TWITTER_CONSUMER_API_KEY", "")
	Twitter.ConsumerAPIKeySecret = getEnvOrDefault("TWITTER_CONSUMER_API_KEY_SECRET", "")
	Twitter.AccessToken = getEnvOrDefault("TWITTER_ACCESS_TOKEN", "")
	Twitter.AccessTokenSecret = getEnvOrDefault("TWITTER_ACCESS_TOKEN_SECRET", "")
}
