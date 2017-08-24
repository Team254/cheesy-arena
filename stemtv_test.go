// Copyright 2016 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublishMatchVideoSplit(t *testing.T) {
	setupTest(t)

	eventSettings.StemTvEventCode = "my_event_code"

	// Mock the STEMtv server.
	stemTvServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/event/api/v1.0/my_event_code/qm254/split/981187501,981187690", r.URL.String())
	}))
	defer stemTvServer.Close()
	stemTvBaseUrl = stemTvServer.URL

	matchStartedTime, _ := time.Parse("2006-01-02 15:04:05 -0700", "2001-02-03 04:05:06 -0400")
	match := &model.Match{Type: "qualification", DisplayName: "254", StartedAt: matchStartedTime}
	scoreDisplayTime, _ := time.Parse("2006-01-02 15:04:05 -0700", "2001-02-03 04:08:00 -0400")
	assert.Nil(t, PublishMatchVideoSplit(match, scoreDisplayTime))
}
