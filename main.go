package main

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/greytabby/newnify/interfaces/controller"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func main() {
	// create router
	ctx := context.Background()
	fsClient, err := createFireStoreClient(ctx)
	rssContoroller := &controller.RSSContoroller{FsClient: fsClient}
	if err != nil {
		logrus.Fatal(err)
	}
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	r.Use(cors.New(config))
	r.GET("/channels", rssContoroller.GetChannels)
	r.POST("/channels", rssContoroller.PostChannel)
	r.GET("/channels/:id/feeds", rssContoroller.GetChannelFeeds)
	r.Run(":7777")
}

func createFireStoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := "greytabby-lab"
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, xerrors.Errorf("Cannot create firestore client: %w", err)
	}
	return client, nil
}
