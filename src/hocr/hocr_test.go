package hocr

import (
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/schollz/cutter"
)

func TestHOCR(t *testing.T) {
	// hocr, err := GetLineDetailsCustomImg("1.hocr", "1.bpm")
	// if err != nil {
	// 	panic(err)
	// }
	// for i, line := range hocr {
	// 	fmt.Println(line.Text)
	// 	f, _ := os.Create("1.jpg")
	// 	line.Img.CopyLineTo(f)
	// 	f.Close()
	// 	if i == 3 {
	// 		break
	// 	}
	// }
	b, err := ioutil.ReadFile("1.hocr")
	if err != nil {
		panic(err)
	}
	hocr, err := Parse(b)
	if err != nil {
		t.Errorf("could not parse: %s", err)
	}
	f, err := os.Open("1.jpg")
	if err != nil {
		log.Fatal("Cannot open file", err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		log.Fatal("Cannot decode image:", err)
	}

	re := regexp.MustCompile("[0-9]+")
	for _, page := range hocr.Pages {
		for li, l := range page.Lines {
			if li != 100 {
				continue
			}
			for _, w := range l.Words {
				startPos := make([]int, 2)
				endPos := make([]int, 2)
				wordtext := ""
				if w.Class != "ocrx_word" {
					continue
				}
				for _, c := range w.Chars {
					if c.Class != "ocrx_cinfo" {
						continue
					}
					wordtext += c.Text
					fmt.Println(c.Title)
					coordinates := re.FindAllString(c.Title, -1)
					fmt.Println(coordinates)
					if startPos[0] == 0 && startPos[1] == 0 {
						startPos[0], _ = strconv.Atoi(coordinates[0])
						startPos[1], _ = strconv.Atoi(coordinates[1])
					}
					endPos[0], _ = strconv.Atoi(coordinates[2])
					endPos[1], _ = strconv.Atoi(coordinates[3])
				}
				wordtext = strings.TrimSpace(wordtext)
				fmt.Println(wordtext)
				fmt.Println(startPos)
				fmt.Println(endPos)
				border := 7
				cImg, err := cutter.Crop(img, cutter.Config{
					Width:   endPos[0] - startPos[0] + border*2,                      // width in pixel or X ratio
					Height:  endPos[1] - startPos[1] + border*2,                      // height in pixel or Y ratio(see Ratio Option below)
					Mode:    cutter.TopLeft,                                          // Accepted Mode: TopLeft, Centered
					Anchor:  image.Point{startPos[0] - border, startPos[1] - border}, // Position of the top left point
					Options: 0,                                                       // Accepted Option: Ratio
				})
				if err != nil {
					log.Fatal(err)
				}
				f2, err := os.Create(wordtext + ".png")
				if err != nil {
					// Handle error
				}
				defer f2.Close()
				err = png.Encode(f2, cImg)
				if err != nil {
					// Handle error
				}

			}
			return
		}
	}
}
