// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Common utility methods for display web routes.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/field"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Returns true if the given required parameters are present; otherwise redirects to the defaults and returns false.
func (web *Web) enforceDisplayConfiguration(w http.ResponseWriter, r *http.Request, defaults map[string]string) bool {
	allPresent := true
	configuration := make(map[string]string)

	// Get display ID and nickname from the query parameters.
	var displayId string
	if displayId = r.URL.Query().Get("displayId"); displayId == "" {
		displayId = web.arena.NextDisplayId()
		allPresent = false
	}
	if nickname := r.URL.Query().Get("nickname"); nickname != "" {
		configuration["nickname"] = nickname
	}

	// Get display-specific fields from the query parameters.
	if defaults != nil {
		for key, defaultValue := range defaults {
			if configuration[key] = r.URL.Query().Get(key); configuration[key] == "" {
				configuration[key] = defaultValue
				allPresent = false
			}
		}
	}

	if !allPresent {
		var builder strings.Builder
		for key, value := range configuration {
			builder.WriteString("&")
			builder.WriteString(url.QueryEscape(key))
			builder.WriteString("=")
			builder.WriteString(url.QueryEscape(value))
		}
		http.Redirect(w, r, fmt.Sprintf("%s?displayId=%s%s", r.URL.Path, displayId, builder.String()), 302)
	}
	return allPresent
}

// Constructs, registers, and returns the display object for the given incoming websocket request.
func (web *Web) registerDisplay(r *http.Request) (*field.Display, error) {
	displayConfig, err := field.DisplayFromUrl(r.URL.Path, r.URL.Query())
	if err != nil {
		return nil, err
	}

	// Extract the source IP address of the request and store it in the display object.
	var ipAddress string
	if ipAddress = r.Header.Get("X-Real-IP"); ipAddress == "" {
		ipAddress = regexp.MustCompile("(.*):\\d+$").FindStringSubmatch(r.RemoteAddr)[1]
	}

	return web.arena.RegisterDisplay(displayConfig, ipAddress), nil
}
