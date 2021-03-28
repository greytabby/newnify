package controller

import (
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type TweetController struct {
	FsClient      *firestore.Client
	TwitterClient *twitter.Client
}

func (ctrl *TweetController) GetHomeTimeline(c *gin.Context) {
	tweets, err := ctrl.getHomeTimeline()
	if err != nil {
		logrus.Error(err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}
	ResponseOK(c, tweets)
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
	list, err := ctrl.postList(ctx, params.ListID)
	if err != nil {
		logrus.Errorf("Registering twitter list failed: %+v", err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}

	ResponseCreated(c, list)
}

func (ctrl *TweetController) GetLists(c *gin.Context) {
	ctx := c.Request.Context()
	lists, err := ctrl.getLists(ctx)
	if err != nil {
		logrus.Errorf("Failed to fetching lists: %+v", err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}

	ResponseOK(c, lists)
}

func (ctrl *TweetController) GetListTimeline(c *gin.Context) {
	listIDStr := c.Param("id")
	listID, err := strconv.ParseInt(listIDStr, 10, 64)
	if err != nil {
		logrus.Infof("Requested ListID is not intger: %v Error: %+v", listIDStr, err)
		ResponseBadRequest(c, "ListID must be integer value")
		return
	}
	tweets, err := ctrl.getListTimeline(c.Request.Context(), listID)
	if err != nil {
		logrus.Error(err)
		ResponseInternalServerError(c, "Internal server error")
		return
	}
	ResponseOK(c, tweets)
}
