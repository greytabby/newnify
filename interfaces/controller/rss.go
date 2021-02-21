package controller

import (
	"context"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CollectionRSSChannels = "RssChannels"
)

type RSSContoroller struct {
	FsClient *firestore.Client
}

type RSSChannelFeeds struct {
	Channel *RSSChannel `json:"channel"`
	Items   []*RSSItem  `json:"items"`
}

type RSSChannel struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	RSSLink     string `json:"rssLink"`
	Description string `json:"description"`
}

type RSSItem struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Link        string     `json:"link"`
	Published   *time.Time `json:"published"`
	GUID        string     `json:"guid"`
	Read        bool       `json:"read"`
}

func (ctrl *RSSContoroller) GetChannelFeeds(c *gin.Context) {
	ctx := c.Request.Context()
	channelID := c.Param("id")
	channelFeeds, err := ctrl.getChannelFeeds(ctx, channelID)
	if err != nil {
		log.Printf("Fetching rss feed failed: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
	}

	c.JSON(http.StatusOK, channelFeeds)
}

func (ctrl *RSSContoroller) GetChannels(c *gin.Context) {
	ctx := c.Request.Context()
	result, err := ctrl.getChannels(ctx)
	if err != nil {
		logrus.Errorf("%+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "Internal server error",
		})
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *RSSContoroller) PostChannel(c *gin.Context) {
	ctx := c.Request.Context()
	var channel RSSChannel
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		logrus.Errorf("BadRequest", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "BadRequest",
		})
		return
	}

	if channel.RSSLink == "" {
		logrus.Errorf("BadRequest", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "RSSLink is required",
		})
		return
	}

	result, err := ctrl.postChannel(ctx, &channel)
	if err != nil {
		logrus.Errorf("Registering rss channel failed: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (ctrl *RSSContoroller) getChannelFeeds(ctx context.Context, channelID string) (*RSSChannelFeeds, error) {
	// channel情報をDBから取得
	snap, err := ctrl.FsClient.Collection(CollectionRSSChannels).Doc(channelID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logrus.Infof("Channel was not found: ChannelID: %v", channelID)
			return nil, nil
		}
	}
	var targetChannel RSSChannel
	err = snap.DataTo(&targetChannel)
	if err != nil {
		return nil, xerrors.Errorf("Cannot bind channel data: %w", err)
	}

	// feedを取得
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(targetChannel.RSSLink, ctx)
	if err != nil {
		return nil, xerrors.Errorf("Error while parsing channel link: %v: %w", targetChannel.Link, err)
	}

	items := feed.Items
	rssItems := make([]*RSSItem, len(items))
	for i, v := range items {
		item := &RSSItem{
			Title:       v.Title,
			Link:        v.Link,
			Description: v.Description,
			Published:   v.PublishedParsed,
			GUID:        v.GUID,
			Read:        false,
		}
		rssItems[i] = item
	}
	// ctrl.FsClient.Collection("test").Add(ctx, channel)
	channelFeeds := &RSSChannelFeeds{
		Channel: &targetChannel,
		Items:   rssItems,
	}
	return channelFeeds, nil
}

func (ctrl *RSSContoroller) postChannel(ctx context.Context, channel *RSSChannel) (*RSSChannel, error) {
	// 有効なRSSチャンネルか確認
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(channel.RSSLink, ctx)
	if err != nil {
		return nil, xerrors.Errorf("Invalid RSS link: %w", err)
	}

	// channel情報をいれる
	guid := xid.New()
	channel.ID = guid.String()
	channel.Link = feed.Link
	channel.Title = feed.Title
	channel.Description = feed.Description

	// DBに保存
	_, err = ctrl.FsClient.Collection(CollectionRSSChannels).Doc(channel.ID).Set(ctx, channel)
	if err != nil {
		return nil, xerrors.Errorf("Write channel to DB was failed: %w", err)
	}

	return channel, nil
}

func (ctrl *RSSContoroller) getChannels(ctx context.Context) ([]*RSSChannel, error) {
	// 登録済みのチャンネルをFirestoreから取得
	docRefs := ctrl.FsClient.Collection(CollectionRSSChannels).DocumentRefs(ctx)
	refs, err := docRefs.GetAll()
	if err != nil {
		return nil, xerrors.Errorf("Cannot read collection of %s: %w", CollectionRSSChannels, err)
	}

	channels := make([]*RSSChannel, 0)
	for _, v := range refs {
		snap, _ := v.Get(ctx)
		var channel RSSChannel
		snap.DataTo(&channel)
		channels = append(channels, &channel)
	}
	return channels, nil
}
