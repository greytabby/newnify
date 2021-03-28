package main

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/greytabby/newnify/appconfig"
	"github.com/greytabby/newnify/interfaces/controller"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func main() {
	appconfig.NewConfig()
	logrus.Infof("%+v", appconfig.App)

	// create router
	ctx := context.Background()
	fsClient, err := createFireStoreClient(ctx)
	rssContoroller := &controller.RSSContoroller{FsClient: fsClient}
	tweetController := &controller.TweetController{FsClient: fsClient, TwitterClient: newTwitterClient()}
	if err != nil {
		logrus.Fatal(err)
	}
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{appconfig.App.AllowOrigin}
	r.Use(cors.New(config))
	r.GET("/rss/channels", rssContoroller.GetChannels)
	r.POST("/rss/channels", rssContoroller.PostChannel)
	r.GET("/rss/channels/:id/feeds", rssContoroller.GetChannelFeeds)
	r.GET("/twitter/hometimeline", tweetController.GetHomeTimeline)
	r.POST("/twitter/lists", tweetController.PostList)
	r.GET("/twitter/lists", tweetController.GetLists)
	r.GET("/twitter/lists/:id/timeline", tweetController.GetListTimeline)
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

func newTwitterClient() *twitter.Client {
	config := oauth1.NewConfig(appconfig.Twitter.ConsumerAPIKey, appconfig.Twitter.ConsumerAPIKeySecret)
	token := oauth1.NewToken(appconfig.Twitter.AccessToken, appconfig.Twitter.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	return twitter.NewClient(httpClient)
}
