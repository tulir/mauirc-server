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

// Package messages contains mauIRC client <-> server messages
package messages

// Preview is an URL preview
type Preview struct {
	Text  *Text  `json:"text,omitempty"`
	Image *Image `json:"image,omitempty"`
}

// ParsePreview parses a Preview object from a generic object
func ParsePreview(obj interface{}) *Preview {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return nil
	}

	var pw Preview

	txt, ok := mp["text"]
	if ok {
		pw.Text = ParseText(txt)
	}

	img, ok := mp["image"]
	if ok {
		pw.Image = ParseImage(img)
	}

	return &pw
}

// Text is some text to preview
type Text struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	SiteName    string `json:"sitename,omitempty"`
}

// ParseText parses a Text object from a generic object
func ParseText(obj interface{}) *Text {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return nil
	}

	var txt Text
	txt.Title, _ = mp["title"].(string)
	txt.Description, _ = mp["description"].(string)
	txt.SiteName, _ = mp["sitename"].(string)
	return &txt
}

// Image is an image to preview
type Image struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
}

// ParseImage parses an Image object from a generic object
func ParseImage(obj interface{}) *Image {
	mp, ok := obj.(map[string]interface{})
	if !ok {
		return nil
	}

	var img Image
	img.URL, _ = mp["url"].(string)
	img.Type, _ = mp["type"].(string)
	img.Width, _ = mp["width"].(uint64)
	img.Height, _ = mp["height"].(uint64)
	return &img
}
