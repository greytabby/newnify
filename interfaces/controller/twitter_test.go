package controller_test

import (
	"testing"

	"github.com/greytabby/newnify/appconfig"
	"github.com/greytabby/newnify/interfaces/controller"
)

func TestFetchHomeTimeLine(t *testing.T) {
	appconfig.NewConfig()
	controller.FetchHomeTimeline()
	t.Log("end")
}
