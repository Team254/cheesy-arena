// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestYouTubeDisplay(t *testing.T) {
	web := setupTestWeb(t)
	for _, captions := range []string{"true", "false"} {
		t.Run(captions, func(t *testing.T) {
			recorder := web.getHttpResponse("/displays/youtube?displayId=1&videoId=liveVideoId&captions=" + captions)
			assert.Equal(t, 200, recorder.Code)
			assert.Contains(t, recorder.Body.String(), "YouTube Stream Display - Untitled Event - Cheesy Arena")
			assert.Contains(t, recorder.Body.String(), "/static/js/youtube_display.js")
		})
	}
}

func TestStreamDisplayDefaults(t *testing.T) {
	web := setupTestWeb(t)
	for _, test := range []struct {
		path     string
		query    string
		expected url.Values
	}{
		{"youtube", "", url.Values{"displayId": {"100"}, "videoId": {""}, "captions": {"false"}}},
		{"youtube", "?displayId=7&nickname=Pit&videoId=liveVideoId", url.Values{
			"displayId": {"7"}, "nickname": {"Pit"}, "videoId": {"liveVideoId"}, "captions": {"false"},
		}},
		{"youtube", "?videoId=liveVideoId&captions=true", url.Values{
			"displayId": {"100"}, "videoId": {"liveVideoId"}, "captions": {"true"},
		}},
		{"youtube", "?videoId=liveVideoId&captions=false", url.Values{
			"displayId": {"100"}, "videoId": {"liveVideoId"}, "captions": {"false"},
		}},
		{"twitch", "", url.Values{"displayId": {"100"}, "channel": {"team254"}, "captions": {"false"}}},
		{"twitch", "?displayId=7&channel=firstinspires", url.Values{
			"displayId": {"7"}, "channel": {"firstinspires"}, "captions": {"false"},
		}},
		{"twitch", "?channel=firstinspires&captions=false", url.Values{
			"displayId": {"100"}, "channel": {"firstinspires"}, "captions": {"false"},
		}},
	} {
		t.Run(test.path+test.query, func(t *testing.T) {
			recorder := web.getHttpResponse("/displays/" + test.path + test.query)
			require.Equal(t, 302, recorder.Code)
			location, err := url.Parse(recorder.Header().Get("Location"))
			require.NoError(t, err)
			assert.Equal(t, "/displays/"+test.path, location.Path)
			assert.Equal(t, test.expected, location.Query())
			assert.Equal(t, 200, web.getHttpResponse(location.String()).Code)
		})
	}
}

func TestYouTubeDisplayWebsocket(t *testing.T) {
	web := setupTestWeb(t)
	server, wsUrl := web.startTestServer()
	defer server.Close()
	conn, _, err := gorillawebsocket.DefaultDialer.Dial(
		wsUrl+"/displays/youtube/websocket?displayId=123&captions=true&videoId=liveVideoId", nil,
	)
	require.NoError(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)
	assert.Equal(t, "/displays/youtube?displayId=123&captions=true&videoId=liveVideoId",
		readWebsocketType(t, ws, "displayConfiguration"))

	// Configuration changes and both reload commands must reach this display, just as they do for Twitch.
	_, err = web.arena.UpdateDisplay(field.DisplayConfiguration{
		Id: "123", Type: field.YouTubeStreamDisplay,
		Configuration: map[string]string{"videoId": "anotherLive", "captions": "false"},
	})
	require.NoError(t, err)
	assert.Equal(t, "/displays/youtube?displayId=123&captions=false&videoId=anotherLive",
		readWebsocketType(t, ws, "displayConfiguration"))
	web.arena.ReloadDisplaysNotifier.NotifyWithMessage("123")
	assert.Equal(t, "123", readWebsocketType(t, ws, "reload"))
	web.arena.ReloadDisplaysNotifier.Notify()
	assert.Nil(t, readWebsocketType(t, ws, "reload"))
}
