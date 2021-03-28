package controller

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"
	"github.com/greytabby/newnify/appconfig"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/xerrors"
)

const (
	TwitterCollection = "twitter"
)

type TweetController struct {
	FsClient *firestore.Client
}

func (ctrl *TweetController) GetHomeTimeLine(c *gin.Context) {
	tweets, err := FetchHomeTimeline()
	if err != nil {
		logrus.Error(err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}
	c.JSON(http.StatusOK, tweets)
}

type PostListParams struct {
	ListID int64 `json:"listId"`
}

func (ctrl *TweetController) PostList(c *gin.Context) {
	var params PostListParams
	err := c.ShouldBindJSON(&params)
	if err != nil {
		logrus.Errorf("BadRequest: %+v", err)
		ResponseBadRequest(c, "Cannot bind request parameter")
		return
	}

	ctx := c.Request.Context()
	list, err := ctrl.AddList(ctx, params.ListID)
	if err != nil {
		logrus.Errorf("Registering twitter list failed: %+v", err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}

	c.JSON(http.StatusCreated, list)
}

func (ctrl *TweetController) GetAllList(c *gin.Context) {
	ctx := c.Request.Context()
	lists, err := ctrl.FetchLists(ctx)
	if err != nil {
		logrus.Errorf("Failed to fetching lists: %+v", err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}

	c.JSON(http.StatusOK, lists)
}

func (ctrl *TweetController) GetListTimeline(c *gin.Context) {
	listIDStr := c.Param("id")
	listID, err := strconv.ParseInt(listIDStr, 10, 64)
	if err != nil {
		logrus.Infof("Requested ListID is not intger: %v Error: %+v", listIDStr, err)
		ResponseBadRequest(c, "ListID must be integer value")
		return
	}
	tweets, err := ctrl.FetchListTimeline(c.Request.Context(), listID)
	if err != nil {
		logrus.Error(err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}
	c.JSON(http.StatusOK, tweets)
}

func FetchHomeTimeline() ([]twitter.Tweet, error) {
	config := oauth1.NewConfig(appconfig.Twitter.ConsumerAPIKey, appconfig.Twitter.ConsumerAPIKeySecret)
	token := oauth1.NewToken(appconfig.Twitter.AccessToken, appconfig.Twitter.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	// verify
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot verify my account: %w", err)
	}

	// get timeline
	homeTImelineParams := &twitter.HomeTimelineParams{
		Count:     1000,
		TweetMode: "extended",
	}
	tweets, _, err := client.Timelines.HomeTimeline(homeTImelineParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot fetch home timeline: %w", err)
	}
	return tweets, nil
}

func (ctrl *TweetController) FetchListTimeline(ctx context.Context, listID int64) ([]twitter.Tweet, error) {
	// config := oauth1.NewConfig(appconfig.Twitter.ConsumerAPIKey, appconfig.Twitter.ConsumerAPIKeySecret)
	// token := oauth1.NewToken(appconfig.Twitter.AccessToken, appconfig.Twitter.AccessTokenSecret)

	config := &clientcredentials.Config{
		ClientID:     appconfig.Twitter.ConsumerAPIKey,
		ClientSecret: appconfig.Twitter.ConsumerAPIKeySecret,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext)
	client := twitter.NewClient(httpClient)

	listStatusesParams := &twitter.ListsStatusesParams{
		ListID: listID,
		Count:  500,
	}
	tweets, _, err := client.Lists.Statuses(listStatusesParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot fetch list timeline. ListID: %v Error: %w", listStatusesParams.ListID, err)
	}
	return tweets, nil
}

func (ctrl *TweetController) AddList(ctx context.Context, listID int64) (*twitter.List, error) {
	config := &clientcredentials.Config{
		ClientID:     appconfig.Twitter.ConsumerAPIKey,
		ClientSecret: appconfig.Twitter.ConsumerAPIKeySecret,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext)
	client := twitter.NewClient(httpClient)

	listsShowParams := &twitter.ListsShowParams{
		ListID: listID,
	}
	list, _, err := client.Lists.Show(listsShowParams)
	if err != nil {
		return nil, xerrors.Errorf("Error: %w", err)
	}
	fmt.Printf("List: %+v\n\n", list)

	if _, err := ctrl.FsClient.Collection(TwitterCollection).Doc("lists").Collection("items").Doc(list.IDStr).Set(ctx, list); err != nil {
		return nil, xerrors.Errorf("Error: %w", err)
	}
	return list, nil
}

func (ctrl *TweetController) FetchLists(ctx context.Context) ([]*twitter.List, error) {
	docRefs := ctrl.FsClient.Collection(TwitterCollection).Doc("lists").Collection("items").DocumentRefs(ctx)
	refs, err := docRefs.GetAll()
	if err != nil {
		return nil, xerrors.Errorf("Cannot read collection of %s: %w", TwitterCollection, err)
	}

	lists := make([]*twitter.List, 0)
	for _, v := range refs {
		snap, _ := v.Get(ctx)
		var list twitter.List
		snap.DataTo(&list)
		lists = append(lists, &list)
	}
	return lists, nil
}
