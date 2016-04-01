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

// Package preview contains URL previewing things
package preview

import (
	"bytes"
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/mvdan/xurls"
	"io/ioutil"
	"net/http"
	"strings"
)

// Preview is an URL preview
type Preview struct {
	Text  *Text  `json:"text,omitempty"`
	Image *Image `json:"image,omitempty"`
}

// Text is some text to preview
type Text struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	SiteName    string `json:"sitename,omitempty"`
}

// Image is an image to preview
type Image struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

// GetPreview gets the preview for the first URL in the given text.
func GetPreview(text string) (*Preview, error) {
	url := xurls.Relaxed.FindString(text)

	if len(url) <= 0 {
		return nil, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ct := http.DetectContentType(data)
	if strings.HasPrefix(ct, "image/") {
		return &Preview{Image: &Image{URL: url, Type: ct}}, nil
	}

	og := opengraph.NewOpenGraph()
	err = og.ProcessHTML(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	var image *Image
	if len(og.Images) > 0 {
		img := og.Images[0]
		if len(img.SecureURL) > 0 {
			image = &Image{URL: img.SecureURL, Type: img.Type, Width: img.Width, Height: img.Height}
		} else {
			image = &Image{URL: img.URL, Type: img.Type, Width: img.Width, Height: img.Height}
		}
	}
	var pwText = &Text{Title: og.Title, Description: og.Description, SiteName: og.SiteName}

	return &Preview{Text: pwText, Image: image}, nil
}
