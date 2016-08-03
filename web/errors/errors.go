// mauIRCd - The IRC bouncer/backend system for mauIRC clients.
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
	InvalidMethod      = WebError{HTTP: http.StatusMethodNotAllowed, Simple: "methodnotallowed", Human: "The request method is not allowed", ExtraInfo: "See the Allow header for a list of allowed headers"}
	InvalidCredentials = WebError{HTTP: http.StatusUnauthorized, Simple: "invalidcredentials", Human: "Invalid username or password"}
	InvalidResetToken  = WebError{HTTP: http.StatusUnauthorized, Simple: "invalidresettoken", Human: "Invalid or expired password reset token"}
	UserNotFound       = WebError{HTTP: http.StatusNotFound, Simple: "usernotfound", Human: "The given email is not in use"}
	NetworkNotFound    = WebError{HTTP: http.StatusNotFound, Simple: "networknotfound", Human: "You don't have a network with the given name"}
	ScriptNotFound     = WebError{HTTP: http.StatusNotFound, Simple: "scriptnotfound", Human: "You don't have a script with the given name"}
	NotAuthenticated   = WebError{HTTP: http.StatusUnauthorized, Simple: "notauthenticated", Human: "You have not logged in", ExtraInfo: "Try logging in using /auth/login"}
	EmailUsed          = WebError{HTTP: http.StatusForbidden, Simple: "emailused", Human: "The given email is already in use"}
	CookieFail         = WebError{HTTP: http.StatusInternalServerError, Simple: "cookiefail", Human: "Failed to find or create the cookie store", ExtraInfo: "Try removing all cookies for this site"}
	BodyNotFound       = WebError{HTTP: http.StatusBadRequest, Simple: "bodynotfound", Human: "The request does not contain a valid body"}
	RequestNotJSON     = WebError{HTTP: http.StatusBadRequest, Simple: "requestnotjson", Human: "The request was not valid JSON"}
	MissingFields      = WebError{HTTP: http.StatusBadRequest, Simple: "missingfields", Human: "The request is missing one or more required fields"}
	FieldFormatting    = WebError{HTTP: http.StatusBadRequest, Simple: "fieldformat", Human: "The request has one or more fields with an invalid format"}
	Internal           = WebError{HTTP: http.StatusInternalServerError, Simple: "internalerror", Human: "An unexpected error occured on the server"}
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
