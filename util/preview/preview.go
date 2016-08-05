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

// Package preview contains URL previewing things
package preview

import (
	"bytes"
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/mvdan/xurls"
	"io/ioutil"
	"maunium.net/go/mauirc-common/messages"
	"net/http"
	"strings"
)

// GetPreview gets the preview for the first URL in the given text.
func GetPreview(text string) (*messages.Preview, error) {
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
		return &messages.Preview{Image: &messages.Image{URL: url, Type: ct}}, nil
	}

	og := opengraph.NewOpenGraph()
	err = og.ProcessHTML(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return &messages.Preview{Text: getText(og), Image: getImage(og)}, nil
}

func getText(og *opengraph.OpenGraph) *messages.Text {
	if len(og.Title) != 0 || len(og.Description) != 0 || len(og.SiteName) != 0 {
		var txt = &messages.Text{Title: og.Title, Description: og.Description, SiteName: og.SiteName}
		if txt.Title == txt.Description {
			txt.Description = ""
		}
		return txt
	}
	return nil
}

func getImage(og *opengraph.OpenGraph) *messages.Image {
	if len(og.Images) > 0 {
		img := og.Images[0]
		if len(img.SecureURL) > 0 {
			return &messages.Image{URL: img.SecureURL, Type: img.Type, Width: img.Width, Height: img.Height}
		}
		return &messages.Image{URL: img.URL, Type: img.Type, Width: img.Width, Height: img.Height}
	}
	return nil
}
