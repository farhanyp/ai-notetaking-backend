package serverutils

import (
	"bytes"
	"io"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractTextFromPdf(reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	readerAt := bytes.NewReader(data)
	content, err := pdf.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return "", err
	}

	styledTexts, err := content.GetStyledTexts()
	if err != nil {
		return "", err
	}

	rows := make(map[int][]pdf.Text)
	var yCoords []int

	for _, st := range styledTexts {
		yKey := int(st.Y)
		if _, ok := rows[yKey]; !ok {
			yCoords = append(yCoords, yKey)
		}
		rows[yKey] = append(rows[yKey], st)
	}

	sort.Slice(yCoords, func(i, j int) bool {
		return yCoords[i] > yCoords[j]
	})

	var buf bytes.Buffer
	for _, y := range yCoords {
		rowTexts := rows[y]

		sort.Slice(rowTexts, func(i, j int) bool {
			return rowTexts[i].X < rowTexts[j].X
		})

		var lineContent strings.Builder
		for _, st := range rowTexts {
			lineContent.WriteString(st.S + " ")
		}

		finalLine := strings.TrimSpace(lineContent.String())
		if finalLine != "" {
			buf.WriteString(finalLine + "\n")
		}
	}

	return buf.String(), nil
}
