package words

import "testing"

func TestWords(t *testing.T) {
	err := Words(Options{
		FileHOCR:     "1.hocr",
		FileImage:    "1.jpg",
		FolderResult: "newspaper",
	})
	if err != nil {
		t.Error(err)
	}
}
