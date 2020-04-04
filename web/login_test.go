// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoginDisplay(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.AdminPassword = "admin"

	// Check that hitting a protected page redirects to the login.
	recorder := web.getHttpResponse("/match_play?p1=v1&p2=v2")
	assert.Equal(t, 307, recorder.Code)
	assert.Equal(t, "/login?redirect=%2Fmatch_play%3Fp1%3Dv1%26p2%3Dv2", recorder.Header().Get("Location"))

	recorder = web.getHttpResponse("/login?redirect=%2Fmatch_play%3Fp1%3Dv1%26p2%3Dv2")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Log In - Untitled Event - Cheesy Arena")

	// Check logging in with the wrong username and right password.
	recorder = web.postHttpResponse("/login?redirect=%2Fmatch_play%3Fp1%3Dv1%26p2%3Dv2",
		"username=blorpy&password=reader")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid login credentials.")

	// Check logging in with the right username and wrong password.
	recorder = web.postHttpResponse("/login?redirect=%2Fmatch_play%3Fp1%3Dv1%26p2%3Dv2",
		"username=admin&password=blorpy")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Invalid login credentials.")

	// Check logging in with the right username and password.
	recorder = web.postHttpResponse("/login?redirect=%2Fmatch_play%3Fp1%3Dv1%26p2%3Dv2",
		"username=admin&password=admin")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/match_play?p1=v1&p2=v2", recorder.Header().Get("Location"))
	cookie := recorder.Header().Get("Set-Cookie")
	assert.Contains(t, cookie, "session_token=")

	// Check that hitting the reader-level protected page works now.
	recorder = web.getHttpResponseWithHeaders("/match_play?p1=v1&p2=v2", map[string]string{"Cookie": cookie})
	assert.Equal(t, 200, recorder.Code)
}
