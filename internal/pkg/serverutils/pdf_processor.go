package serverutils

import (
	"bytes"
	"io"

	"github.com/ledongthuc/pdf"
)

func ExtractTextFromPdf(reader io.Reader) (string, error) {
	// Library ini butuh io.ReaderAt, jadi kita baca dulu ke buffer
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	readerAt := bytes.NewReader(data)
	content, err := pdf.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	for i := 1; i <= content.NumPage(); i++ {
		p := content.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, _ := p.GetPlainText(nil)
		buf.WriteString(text)
	}

	return buf.String(), nil
}
