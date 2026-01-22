package serverutils

import (
	"bytes"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PdfPage struct {
	PageNumber int
	Content    string
}

func ExtractTextPerPage(reader io.Reader) ([]PdfPage, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	readerAt := bytes.NewReader(data)
	content, err := pdf.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return nil, err
	}

	var pages []PdfPage

	for pageIndex := 1; pageIndex <= content.NumPage(); pageIndex++ {
		page := content.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return nil, err
		}

		clean := strings.TrimSpace(text)
		if clean != "" {
			pages = append(pages, PdfPage{
				PageNumber: pageIndex,
				Content:    clean,
			})
		}
	}

	return pages, nil
}
