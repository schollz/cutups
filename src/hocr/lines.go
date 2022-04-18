// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package hocr

// TODO: Parse line name to zero pad line numbers, so they can
//       be sorted easily

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/schollz/cutups/src/line"
)

// Returns the image path for a page from a ocr_page title
func imagePathFromTitle(s string) (string, error) {
	re, err := regexp.Compile(`image ["']([^"']+)["']`)
	if err != nil {
		return "", err
	}
	m := re.FindStringSubmatch(s)
	return m[1], nil
}

// LineText extracts the text from an OcrLine
func LineText(l OcrLine) string {
	linetext := ""

	linetext = l.Text
	if noText(linetext) {
		linetext = ""
		for _, w := range l.Words {
			if w.Class != "ocrx_word" {
				continue
			}
			linetext += w.Text + " "
		}
	}
	if noText(linetext) {
		linetext = ""
		for _, w := range l.Words {
			if w.Class != "ocrx_word" {
				continue
			}
			for _, c := range w.Chars {
				if c.Class != "ocrx_cinfo" {
					continue
				}
				linetext += c.Text
			}
			linetext += " "
		}
	}
	linetext = strings.TrimRight(linetext, " ")
	return linetext
}

// parseLineDetails parses a Hocr struct into a line.Details
// struct, including extracted image segments for each line.
// The image location is taken from imgPath, which can either
// be imagePathFromTitle (see above) which loads the image
// path embedded in the title attribute of a hocr page, or
// a custom handler.
func parseLineDetails(h Hocr, dir string, imgPath func(string) (string, error)) (line.Details, error) {
	lines := make(line.Details, 0)

	for _, p := range h.Pages {
		imgpath, err := imgPath(p.Title)
		if err != nil {
			return lines, err
		}
		imgpath = filepath.Join(dir, filepath.Base(imgpath))

		var img image.Image
		var gray *image.Gray
		pngf, err := os.Open(imgpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: error opening image %s: %v\n", imgpath, err)
		}
		defer pngf.Close()
		img, _, err = image.Decode(pngf)
		if err == nil {
			b := img.Bounds()
			gray = image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
			draw.Draw(gray, b, img, b.Min, draw.Src)
		}

		for _, l := range p.Lines {
			totalconf := float64(0)
			num := 0
			for _, w := range l.Words {
				c, err := wordConf(w.Title)
				if err != nil {
					return lines, err
				}
				num++
				totalconf += c
			}

			coords, err := BoxCoords(l.Title)
			if err != nil {
				return lines, err
			}

			var ln line.Detail
			ln.Name = l.Id
			ln.Avgconf = (totalconf / float64(num)) / 100
			ln.Text = LineText(l)
			imgpath, err := imgPath(p.Title)
			if err != nil {
				return lines, err
			}
			ln.OcrName = strings.TrimSuffix(filepath.Base(imgpath), ".png")
			if gray != nil {
				var imgd line.ImgDirect
				imgd.Img = gray.SubImage(image.Rect(coords[0], coords[1], coords[2], coords[3]))
				ln.Img = imgd
			}
			lines = append(lines, ln)
		}
		pngf.Close()
	}
	return lines, nil
}

// GetLineDetails parses a hocr file and returns a corresponding
// line.Details, including image extracts for each line
func GetLineDetails(hocrfn string) (line.Details, error) {
	var newlines line.Details

	file, err := ioutil.ReadFile(hocrfn)
	if err != nil {
		return newlines, err
	}

	h, err := Parse(file)
	if err != nil {
		return newlines, err
	}

	return parseLineDetails(h, filepath.Dir(hocrfn), imagePathFromTitle)
}

// GetLineDetailsCustomImg is a variant of GetLineDetails that
// uses a provided image path for line image extracts, rather
// than the image name embedded in the .hocr
func GetLineDetailsCustomImg(hocrfn string, imgfn string) (line.Details, error) {
	var newlines line.Details

	file, err := ioutil.ReadFile(hocrfn)
	if err != nil {
		return newlines, err
	}

	h, err := Parse(file)
	if err != nil {
		return newlines, err
	}

	return parseLineDetails(h, filepath.Dir(hocrfn), func(s string) (string, error) { return imgfn, nil })
}

// GetLineBasics parses a hocr file and returns a corresponding
// line.Details, without any image extracts
func GetLineBasics(hocrfn string) (line.Details, error) {
	var newlines line.Details

	file, err := ioutil.ReadFile(hocrfn)
	if err != nil {
		return newlines, err
	}

	h, err := Parse(file)
	if err != nil {
		return newlines, err
	}

	return parseLineDetails(h, filepath.Dir(hocrfn), imagePathFromTitle)
}
