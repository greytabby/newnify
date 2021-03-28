package controller

import (
	"log"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RSSContoroller struct {
	FsClient *firestore.Client
}

func (ctrl *RSSContoroller) GetChannelFeeds(c *gin.Context) {
	ctx := c.Request.Context()
	channelID := c.Param("id")
	channelFeeds, err := ctrl.getChannelFeeds(ctx, channelID)
	if err != nil {
		log.Printf("Fetching rss feed failed: %+v", err)
		ResponseInternalServerError(c, "Internal server error")
	}

	ResponseOK(c, channelFeeds)
}

func (ctrl *RSSContoroller) GetChannels(c *gin.Context) {
	ctx := c.Request.Context()
	channels, err := ctrl.getChannels(ctx)
	if err != nil {
		logrus.Errorf("%+v", err)
		ResponseInternalServerError(c, "Internal server error")
	}

	ResponseOK(c, channels)
}

func (ctrl *RSSContoroller) PostChannel(c *gin.Context) {
	ctx := c.Request.Context()
	var channel RSSChannel
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		logrus.Errorf("BadRequest: %+v", err)
		ResponseBadRequest(c, "Cannot bind request parameter")
		return
	}

	if channel.RSSLink == "" {
		logrus.Errorf("BadRequest: %+v", err)
		ResponseBadRequest(c, "RSSLink is required")
		return
	}

	result, err := ctrl.postChannel(ctx, &channel)
	if err != nil {
		logrus.Errorf("Registering rss channel failed: %+v", err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}

	ResponseCreated(c, result)
}
