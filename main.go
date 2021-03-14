package main

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/greytabby/newnify/appconfig"
	"github.com/greytabby/newnify/interfaces/controller"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func main() {
	appconfig.NewConfig()

	// create router
	ctx := context.Background()
	fsClient, err := createFireStoreClient(ctx)
	rssContoroller := &controller.RSSContoroller{FsClient: fsClient}
	tweetController := &controller.TweetController{FsClient: fsClient}
	if err != nil {
		logrus.Fatal(err)
	}
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{appconfig.App.AllowOrigin}
	r.Use(cors.New(config))
	r.GET("/channels", rssContoroller.GetChannels)
	r.POST("/channels", rssContoroller.PostChannel)
	r.GET("/channels/:id/feeds", rssContoroller.GetChannelFeeds)
	r.GET("/tweets", tweetController.GetHomeTimeLine)
	r.Run(":" + appconfig.App.Port)
}

func createFireStoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := "greytabby-lab"
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, xerrors.Errorf("Cannot create firestore client: %w", err)
	}
	return client, nil
}
