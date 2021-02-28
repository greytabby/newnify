package main

import (
	"context"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/greytabby/newnify/interfaces/controller"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

type Config struct {
	allowOrigin string
	port        string
}

var appConfig = Config{
	allowOrigin: "http://localhost:3000",
	port:        "7777",
}

func getEnvOrDefault(name string, defaultValue string) string {
	envValue := os.Getenv(name)
	if envValue == "" {
		return defaultValue
	}
	return envValue
}

func loadConfig() {
	appConfig.allowOrigin = getEnvOrDefault("ALLOW_ORIGIN", "http://localhost:3000")
	appConfig.port = getEnvOrDefault("PORT", "7777")
}

func main() {
	loadConfig()

	// create router
	ctx := context.Background()
	fsClient, err := createFireStoreClient(ctx)
	rssContoroller := &controller.RSSContoroller{FsClient: fsClient}
	if err != nil {
		logrus.Fatal(err)
	}
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{appConfig.allowOrigin}
	r.Use(cors.New(config))
	r.GET("/channels", rssContoroller.GetChannels)
	r.POST("/channels", rssContoroller.PostChannel)
	r.GET("/channels/:id/feeds", rssContoroller.GetChannelFeeds)
	r.Run(":" + appConfig.port)
}

func createFireStoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := "greytabby-lab"
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, xerrors.Errorf("Cannot create firestore client: %w", err)
	}
	return client, nil
}
