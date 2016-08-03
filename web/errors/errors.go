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
	"net/http"
)

type webError struct {
	HTTP      int    `json:"http"`
	Simple    string `json:"error"`
	Human     string `json:"error-humanreadable"`
	ExtraInfo string `json:"error-extrainfo,omitempty"`
}

// Web errors
var (
	InvalidMethod      = webError{HTTP: http.StatusMethodNotAllowed, Simple: "methodnotallowed", Human: "The request method is not allowed", ExtraInfo: "See the Allow header for a list of allowed headers"}
	InvalidCredentials = webError{HTTP: http.StatusUnauthorized, Simple: "invalidcredentials", Human: "Invalid username or password"}
	NotAuthenticated   = webError{HTTP: http.StatusUnauthorized, Simple: "notauthenticated", Human: "You have not logged in", ExtraInfo: "Try logging in using /auth/login"}
	EmailUsed          = webError{HTTP: http.StatusForbidden, Simple: "emailused", Human: "The given email is already in use"}
	CookieFail         = webError{HTTP: http.StatusInternalServerError, Simple: "cookiefail", Human: "Failed to find or create the cookie store", ExtraInfo: "Try removing all cookies for this site"}
	BodyNotFound       = webError{HTTP: http.StatusBadRequest, Simple: "bodynotfound", Human: "The request does not contain a valid body"}
	RequestNotJSON     = webError{HTTP: http.StatusBadRequest, Simple: "requestnotjson", Human: "The request was not valid JSON"}
	MissingFields      = webError{HTTP: http.StatusBadRequest, Simple: "missingfields", Human: "The request is missing one or more required fields"}
	FieldFormatting    = webError{HTTP: http.StatusBadRequest, Simple: "fieldformat", Human: "The request has one or more fields with an invalid format"}
	Internal           = webError{HTTP: http.StatusInternalServerError, Simple: "internalerror", Human: "An unexpected error occured on the server"}
)

// Write a webError to a http.ResponseWriter
func Write(w http.ResponseWriter, err webError) error {
	enc := json.NewEncoder(w)
	w.WriteHeader(err.HTTP)
	return enc.Encode(err)
}
