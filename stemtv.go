// Copyright 2016 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for publishing match video split information to STEMtv.

package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	preMatchPaddingSec         = 5
	postScoreDisplayPaddingSec = 10
)

var stemTvBaseUrl = "http://stemtv.io"

func PublishMatchVideoSplit(match *Match, scoreDisplayTime time.Time) error {
	url := fmt.Sprintf("%s/event/api/v1.0/%s/%s/split/%d,%d", stemTvBaseUrl, eventSettings.StemTvEventCode,
		match.TbaCode(), match.StartedAt.Unix()-preMatchPaddingSec,
		scoreDisplayTime.Unix()+postScoreDisplayPaddingSec)
	_, err := http.Get(url)
	return err
}
