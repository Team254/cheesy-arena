// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoginDisplay(t *testing.T) {
	web := setupTestWeb(t)
	web.arena.EventSettings.ReaderPassword = "reader"
	web.arena.EventSettings.AdminPassword = "admin"

	// Check that hitting a reader-level protected page redirects to the login.
	recorder := web.getHttpResponse("/api/alliances")
	assert.Equal(t, 307, recorder.Code)
	assert.Equal(t, "/login?redirect=/api/alliances", recorder.Header().Get("Location"))

	recorder = web.getHttpResponse("/login?redirect=/api/alliances")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Log In - Untitled Event - Cheesy Arena")

	// Check logging in with the wrong username and right password.
	recorder = web.postHttpResponse("/login?redirect=/api/alliances", "username=blorpy&password=reader")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Bad username or password")

	// Check logging in with the right username and wrong password.
	recorder = web.postHttpResponse("/login?redirect=/api/alliances", "username=reader&password=blorpy")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Bad username or password")

	// Check logging in with the right username and password.
	recorder = web.postHttpResponse("/login?redirect=/api/alliances", "username=reader&password=reader")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/api/alliances", recorder.Header().Get("Location"))
	cookie := recorder.Header().Get("Set-Cookie")
	assert.Contains(t, cookie, "Authorization=")

	// Check that hitting the reader-level protected page works now.
	recorder = web.getHttpResponseWithHeaders("/api/alliances", map[string]string{"Cookie": cookie})
	assert.Equal(t, 200, recorder.Code)

	// Check that hitting a admin-level protected at a higher level requires a different login.
	recorder = web.getHttpResponseWithHeaders("/match_play", map[string]string{"Cookie": cookie})
	assert.Equal(t, 307, recorder.Code)
	assert.Equal(t, "/login?redirect=/match_play", recorder.Header().Get("Location"))
	recorder = web.getHttpResponseWithHeaders("/login?redirect=/match_play", map[string]string{"Cookie": cookie})
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "insufficient privileges")
	recorder = web.postHttpResponse("/login?redirect=/match_play", "username=admin&password=admin")
	assert.Equal(t, 303, recorder.Code)
	assert.Equal(t, "/match_play", recorder.Header().Get("Location"))
	cookie = recorder.Header().Get("Set-Cookie")
	assert.Contains(t, cookie, "Authorization=")
	recorder = web.getHttpResponseWithHeaders("/match_play", map[string]string{"Cookie": cookie})
	assert.Equal(t, 200, recorder.Code)

	// Check that the admin user also has access to the reader-level pages.
	recorder = web.getHttpResponseWithHeaders("/api/alliances", map[string]string{"Cookie": cookie})
	assert.Equal(t, 200, recorder.Code)
}
