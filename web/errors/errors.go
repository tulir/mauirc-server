// mauIRC-server - The IRC bouncer/backend system for mauIRC clients.
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package errors contains the web errors
package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WebError is a wrapper for errors intended to be sent to HTTP clients
type WebError struct {
	HTTP      int    `json:"http"`
	Simple    string `json:"error"`
	Human     string `json:"error-humanreadable"`
	ExtraInfo string `json:"error-extrainfo,omitempty"`
}

// Web errors
var (
	InvalidMethod      = Create(http.StatusMethodNotAllowed, "methodnotallowed", "The request method is not allowed", "See the Allow header for a list of allowed headers")
	InvalidCredentials = Create(http.StatusUnauthorized, "invalidcredentials", "Invalid username or password", "")
	InvalidResetToken  = Create(http.StatusUnauthorized, "invalidresettoken", "Invalid or expired password reset token", "")
	UserNotFound       = Create(http.StatusNotFound, "usernotfound", "The given email is not in use", "")
	NetworkNotFound    = Create(http.StatusNotFound, "networknotfound", "You don't have a network with the given name", "")
	ScriptNotFound     = Create(http.StatusNotFound, "scriptnotfound", "You don't have a script with the given name", "")
	NotAuthenticated   = Create(http.StatusUnauthorized, "notauthenticated", "You have not logged in", "Try logging in using /auth/login")
	EmailUsed          = Create(http.StatusForbidden, "emailused", "The given email is already in use", "")
	CookieFail         = Create(http.StatusInternalServerError, "cookiefail", "Failed to find or create the cookie store", "Try removing all cookies for this site")
	BodyNotFound       = Create(http.StatusBadRequest, "bodynotfound", "The request does not contain a valid body", "")
	InvalidBodyFormat  = Create(http.StatusBadRequest, "invalidbodyformat", "The request was in an invalid format", "")
	RequestNotJSON     = Create(http.StatusBadRequest, "requestnotjson", "The request was not valid JSON", "")
	MissingFields      = Create(http.StatusBadRequest, "missingfields", "The request is missing one or more required fields", "")
	FieldFormatting    = Create(http.StatusBadRequest, "fieldformat", "The request has one or more fields with an invalid format", "")
	MailerDisabled     = Create(http.StatusForbidden, "mailerdisabled", "The mailing system is disabled", "No actions that require sending mails can be completed")
	Internal           = Create(http.StatusInternalServerError, "internalerror", "An unexpected error occured on the server", "")
)

// Create a custom error
func Create(status int, simple, human, extra string) WebError {
	return WebError{HTTP: status, Simple: simple, Human: human, ExtraInfo: extra}
}

func (err WebError) Error() string {
	if len(err.ExtraInfo) > 0 {
		return fmt.Sprintf("%s: %s. %s (HTTP %d)", err.Simple, err.Human, err.ExtraInfo, err.HTTP)
	}
	return fmt.Sprintf("%s: %s (HTTP %d)", err.Simple, err.Human, err.HTTP)
}

// Write a WebError to a http.ResponseWriter
func Write(w http.ResponseWriter, err WebError) error {
	enc := json.NewEncoder(w)
	w.WriteHeader(err.HTTP)
	return enc.Encode(err)
}
