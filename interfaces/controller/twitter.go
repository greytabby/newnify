package controller

import (
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"
	"github.com/greytabby/newnify/appconfig"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
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
		Count: 1000,
	}
	tweets, _, err := client.Timelines.HomeTimeline(homeTImelineParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot fetch home timeline: %w", err)
	}
	return tweets, nil
}
