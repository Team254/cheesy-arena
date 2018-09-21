// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for authenticating with the server.

package web

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"net/http"
)

// Shows the login form.
func (web *Web) loginHandler(w http.ResponseWriter, r *http.Request) {
	var errorMessage string
	if username := web.cookieAuth.Authorize(r); username != "" {
		// If redirected here but already logged in, the user must have insufficient privileges; show a useful message.
		errorMessage = fmt.Sprintf("User '%s' has insufficient privileges for the requested page. Try logging in as a"+
			" different user.", username)
	}
	web.renderLogin(w, r, errorMessage)
}

// Processes the login request.
func (web *Web) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := web.cookieAuth.Login(w, r.PostFormValue("username"), r.PostFormValue("password")); err != nil {
		web.renderLogin(w, r, err.Error())
		return
	}

	redirectUrl := r.URL.Query().Get("redirect")
	if redirectUrl == "" {
		redirectUrl = "/"
	}
	http.Redirect(w, r, redirectUrl, 303)
}

func (web *Web) renderLogin(w http.ResponseWriter, r *http.Request, errorMessage string) {
	template, err := web.parseFiles("templates/login.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		ErrorMessage string
	}{web.arena.EventSettings, errorMessage}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
