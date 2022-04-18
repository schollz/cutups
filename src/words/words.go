package words

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/disintegration/gift"
	"github.com/flytam/filenamify"
	"github.com/schollz/cutter"
	"github.com/schollz/cutups/src/hocr"
	log "github.com/schollz/logger"
	"github.com/schollz/progressbar/v3"
)

type Options struct {
	FileHOCR     string
	FileImage    string
	FolderResult string
	Border       int
	Threshold    float32
}

func Words(o Options) (err error) {
	if o.Threshold == 0 {
		o.Threshold = 80
	}
	if o.Border == 0 {
		o.Border = 7
	}

	os.MkdirAll(o.FolderResult, 0644)
	// open hocr file
	b, err := ioutil.ReadFile(o.FileHOCR)
	if err != nil {
		log.Error(err)
		return
	}
	// parse hocr file
	hocr, err := hocr.Parse(b)
	if err != nil {
		log.Error(err)
		return
	}

	// open up image
	f, err := os.Open(o.FileImage)
	if err != nil {
		log.Error(err)
		return
	}
	img, _, err := image.Decode(f)
	if err != nil {
		log.Error(err)
		return
	}

	// compile regex
	re := regexp.MustCompile("[0-9]+")

	// compile image filter
	g := gift.New(
		gift.Grayscale(),
		gift.Contrast(10),
		gift.Threshold(o.Threshold),
		gift.ColorFunc(func(r0, g0, b0, a0 float32) (r, g, b, a float32) {
			if r0 == 0 {
				return 0, 0, 0, 1
			} else {
				return 0, 0, 0, 0
			}
		}),
	)

	wordCount := 0
	for _, page := range hocr.Pages {
		for _, l := range page.Lines {
			wordCount += len(l.Words)
		}
	}

	bar := progressbar.Default(int64(wordCount))
	// go through each page
	for _, page := range hocr.Pages {
		for _, l := range page.Lines {
			for _, w := range l.Words {
				bar.Add(1)
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
				validName, _ := filenamify.Filenamify(wordtext, filenamify.Options{})
				if validName != wordtext {
					log.Debugf("%s != %s", wordtext, validName)
					continue
				}
				fmt.Println(wordtext)
				fmt.Println(startPos)
				fmt.Println(endPos)
				border := o.Border
				var cImg image.Image
				cImg, err = cutter.Crop(img, cutter.Config{
					Width:   endPos[0] - startPos[0] + border*2,                      // width in pixel or X ratio
					Height:  endPos[1] - startPos[1] + border*2,                      // height in pixel or Y ratio(see Ratio Option below)
					Mode:    cutter.TopLeft,                                          // Accepted Mode: TopLeft, Centered
					Anchor:  image.Point{startPos[0] - border, startPos[1] - border}, // Position of the top left point
					Options: 0,                                                       // Accepted Option: Ratio
				})
				if err != nil {
					log.Error(err)
					return
				}

				dst := image.NewRGBA(g.Bounds(cImg.Bounds()))
				g.Draw(dst, cImg)
				var f2 *os.File
				filei := 1
				for ii := 1; ii < 100; ii++ {
					if _, err = os.Stat(path.Join(o.FolderResult, fmt.Sprintf("%s_%d.png", wordtext, ii))); errors.Is(err, os.ErrNotExist) {
						filei = ii
						break
					}
				}

				f2, err = os.Create(path.Join(o.FolderResult, fmt.Sprintf("%s_%d.png", wordtext, filei)))
				if err != nil {
					log.Error(err)
					return
				}
				defer f2.Close()
				err = png.Encode(f2, dst)
				if err != nil {
					log.Error(err)
					return
				}
			}
		}
	}

	return
}
