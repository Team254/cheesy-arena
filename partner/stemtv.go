// Copyright 2016 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for publishing match video split information to STEMtv.

package partner

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"net/http"
	"time"
)

type StemTvClient struct {
	BaseUrl   string
	eventCode string
}

const (
	stemTvBaseUrl              = "http://52.21.72.74"
	preMatchPaddingSec         = 5
	postScoreDisplayPaddingSec = 10
)

func NewStemTvClient(eventCode string) *StemTvClient {
	return &StemTvClient{stemTvBaseUrl, eventCode}
}

func (client *StemTvClient) PublishMatchVideoSplit(match *model.Match, scoreDisplayTime time.Time) error {
	url := fmt.Sprintf("%s/event/api/v1.0/%s/%s/split/%d,%d", client.BaseUrl, client.eventCode, match.TbaCode(),
		match.StartedAt.Unix()-preMatchPaddingSec, scoreDisplayTime.Unix()+postScoreDisplayPaddingSec)
	_, err := http.Get(url)
	return err
}
