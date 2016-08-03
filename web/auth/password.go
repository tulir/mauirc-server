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

// Package auth contains the authentication system
package auth

import (
	"encoding/json"
	"fmt"
	"maunium.net/go/mauircd/web/errors"
	"maunium.net/go/mauircd/web/util"
	"net/http"
)

type passwordResetForm struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// PasswordReset HTTP handler
func PasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		errors.Write(w, errors.InvalidMethod)
		return
	}

	dec := json.NewDecoder(r.Body)
	var prf passwordResetForm
	err := dec.Decode(&prf)

	if err != nil || len(prf.Email) == 0 || len(prf.Password) == 0 || len(prf.Token) == 0 {
		errors.Write(w, errors.MissingFields)
		return
	}

	user := config.GetUser(prf.Email)
	if user == nil {
		errors.Write(w, errors.UserNotFound)
		return
	} else if !user.CheckResetToken(prf.Token) {
		errors.Write(w, errors.InvalidResetToken)
		return
	}

	user.SetPassword(prf.Password)
	w.WriteHeader(http.StatusOK)
}

type passwordForgotForm struct {
	Email string `json:"email"`
}

// PasswordForgot HTTP handler
func PasswordForgot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		errors.Write(w, errors.InvalidMethod)
		return
	}

	dec := json.NewDecoder(r.Body)
	var pff passwordForgotForm
	err := dec.Decode(&pff)

	if err != nil || len(pff.Email) == 0 {
		errors.Write(w, errors.MissingFields)
		return
	}

	user := config.GetUser(pff.Email)
	if user == nil {
		errors.Write(w, errors.UserNotFound)
		return
	}

	token, time := user.NewResetToken()

	config.GetMail().Send(user.GetEmail(), "mauIRC password reset", fmt.Sprintf(
		`Someone from %[1]s has requested a password reset for your mauIRC account.

		If this was you, you may use the following link to reset your password:
		<a href="%[2]s/resetpassword.html?%[3]s">%[2]s/resetpassword.html?%[3]s</a>

		If you did not request this, it is safe to ignore.
		The person requesting this was not granted access to any personal information.
		The link will expire in 30 minutes (%[4]s).`,
		util.GetIP(r), config.GetExternalAddr(), token, time.Format("15:04:05 02.01.2006 UTC-07")))
}

type passwordChangeForm struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// PasswordChange HTTP handler
func PasswordChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Add("Allow", http.MethodPost)
		errors.Write(w, errors.InvalidMethod)
		return
	}

	authd, user := Check(w, r)
	if !authd {
		errors.Write(w, errors.NotAuthenticated)
		return
	}

	dec := json.NewDecoder(r.Body)
	var pcf passwordChangeForm
	err := dec.Decode(&pcf)

	if err != nil || len(pcf.Old) == 0 || len(pcf.New) == 0 {
		errors.Write(w, errors.MissingFields)
		return
	}

	if !user.CheckPassword(pcf.Old) {
		log.Debugf("%s tried to change password of %s with the wrong password\n", util.GetIP(r), user.GetEmail())
		errors.Write(w, errors.InvalidCredentials)
		return
	}

	err = user.SetPassword(pcf.New)
	if err != nil {
		log.Errorf("%s failed to change password of %s: %s", util.GetIP(r), user.GetEmail(), err)
		errors.Write(w, errors.Internal)
		return
	}

	log.Debugf("%s changed the password of %s\n", util.GetIP(r), user.GetEmail())
	w.WriteHeader(http.StatusOK)
}
