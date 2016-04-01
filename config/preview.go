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

// Package config contains configurations
package config

import (
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/mvdan/xurls"
	"net/http"
)

// GetPreview gets the preview for the first URL in the given text.
func GetPreview(text string) (*opengraph.OpenGraph, error) {
	url := xurls.Relaxed.FindString(text)

	if len(url) <= 0 {
		return nil, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	og := opengraph.NewOpenGraph()
	err = og.ProcessHTML(resp.Body)
	if err != nil {
		return nil, err
	}

	return og, nil
}
