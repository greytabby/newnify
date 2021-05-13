package controller

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/greytabby/ogp"
	"github.com/mmcdole/gofeed"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CollectionRSSChannels = "RssChannels"
	CollectionRSSItems    = "RssItems"
)

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
	ImageURL    string     `json:"imageURL"`
	ImageWidth  string     `json:"imageWidth"`
	ImageHeight string     `json:"imageHeight"`
	ImageAlt    string     `json:"imageAlt"`
	DocID       string
}

func (ctrl *RSSContoroller) collectAndSaveFeeds(ctx context.Context) error {
	channels, err := ctrl.getChannels(ctx)
	if err != nil {
		logrus.Errorf("Cannot get channels data")
		return xerrors.New("Cannot get channels data")
	}

	wg := sync.WaitGroup{}

	collectFn := func(channel *RSSChannel) {
		defer wg.Done()
		fp := gofeed.NewParser()
		feed, err := fp.ParseURLWithContext(channel.RSSLink, ctx)
		if err != nil {
			logrus.Errorf("Error while parsing channel link: %v: %w", channel.Link, err)
			return
		}

		rssItems := make([]*RSSItem, 0)
		for _, item := range feed.Items {
			resp, err := http.Get(item.Link)
			if err != nil {
				logrus.Errorf("Item link access failure: %w", err)
				return
			}
			document, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Errorf("Cannot read response body: %w", err)
				return
			}
			og, err := ogp.Parse(document)
			if err != nil {
				logrus.Infof("ogp parse error: %w", err)
			}

			hash := sha1.New()
			io.WriteString(hash, item.Link)
			rssItem := &RSSItem{
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				Published:   item.PublishedParsed,
				GUID:        item.GUID,
				DocID:       hex.EncodeToString(hash.Sum(nil)),
				Read:        false,
			}

			if og != nil && len(og.Images) != 0 {
				image := og.Images[0]
				rssItem.ImageURL = image.URL
				rssItem.ImageWidth = image.Width
				rssItem.ImageHeight = image.Height
				rssItem.ImageAlt = image.Alt
			}
			rssItems = append(rssItems, rssItem)
		}

		batch := ctrl.FsClient.Batch()
		for _, i := range rssItems {
			logrus.Info("DocID:", i.DocID, "Title", i.Title, i.GUID)
			doc := ctrl.FsClient.Collection(CollectionRSSChannels).Doc(channel.ID).Collection(CollectionRSSItems).Doc(i.DocID)
			batch.Set(doc, i)
		}

		_, err = batch.Commit(ctx)
		if err != nil {
			logrus.Errorf("RssFeeds write error: %+v", err)
			return
		}

		return
	}

	for _, channel := range channels {
		wg.Add(1)
		go collectFn(channel)
	}

	wg.Wait()
	return nil
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
