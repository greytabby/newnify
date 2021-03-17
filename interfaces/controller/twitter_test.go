package controller_test

import (
	"testing"

	"github.com/greytabby/newnify/appconfig"
	"github.com/greytabby/newnify/interfaces/controller"
)

func TestFetchHomeTimeLine(t *testing.T) {
	appconfig.NewConfig()
	tweets, _ := controller.FetchHomeTimeline()
	t.Logf("%+v", tweets[0].User)
	t.Log("end")
}
