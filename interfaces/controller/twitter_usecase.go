package controller

import (
	"context"
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/xerrors"
)

const (
	TwitterCollection          = "twitter"
	TwitterListDoc             = "lists"
	TwitterListItemsCollection = "items"
)

func (ctrl *TweetController) getHomeTimeline() ([]twitter.Tweet, error) {
	// verify
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	_, _, err := ctrl.TwitterClient.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot verify my account: %w", err)
	}

	// get timeline
	homeTImelineParams := &twitter.HomeTimelineParams{
		Count:     1000,
		TweetMode: "extended",
	}
	tweets, _, err := ctrl.TwitterClient.Timelines.HomeTimeline(homeTImelineParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot fetch home timeline: %w", err)
	}
	return tweets, nil
}

func (ctrl *TweetController) getListTimeline(ctx context.Context, listID int64) ([]twitter.Tweet, error) {
	listStatusesParams := &twitter.ListsStatusesParams{
		ListID: listID,
		Count:  500,
	}
	tweets, _, err := ctrl.TwitterClient.Lists.Statuses(listStatusesParams)
	if err != nil {
		return nil, xerrors.Errorf("Cannot fetch list timeline. ListID: %v Error: %w", listStatusesParams.ListID, err)
	}
	return tweets, nil
}

func (ctrl *TweetController) postList(ctx context.Context, listID int64) (*twitter.List, error) {
	listsShowParams := &twitter.ListsShowParams{
		ListID: listID,
	}
	list, _, err := ctrl.TwitterClient.Lists.Show(listsShowParams)
	if err != nil {
		return nil, xerrors.Errorf("Error: %w", err)
	}
	fmt.Printf("List: %+v\n\n", list)

	if _, err := ctrl.FsClient.Collection(TwitterCollection).Doc(TwitterListDoc).Collection(TwitterListItemsCollection).Doc(list.IDStr).Set(ctx, list); err != nil {
		return nil, xerrors.Errorf("Error: %w", err)
	}
	return list, nil
}

func (ctrl *TweetController) getLists(ctx context.Context) ([]*twitter.List, error) {
	docRefs := ctrl.FsClient.Collection(TwitterCollection).Doc(TwitterListDoc).Collection(TwitterListItemsCollection).DocumentRefs(ctx)
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
