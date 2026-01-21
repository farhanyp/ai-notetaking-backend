package chunking

import (
	"strings"

	"github.com/tmc/langchaingo/schema"
)

type PdfPage struct {
	PageNumber int
	Content    string
}

func ChunkPdfPage(
	page PdfPage,
	maxChunkSize int,
) []schema.Document {

	var docs []schema.Document

	paragraphs := strings.Split(page.Content, "\n\n")

	var buffer strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		if buffer.Len()+len(p) > maxChunkSize {
			docs = append(docs, schema.Document{
				PageContent: buffer.String(),
				Metadata: map[string]any{
					"page": page.PageNumber,
				},
			})
			buffer.Reset()
		}

		buffer.WriteString(p + "\n\n")
	}

	if buffer.Len() > 0 {
		docs = append(docs, schema.Document{
			PageContent: buffer.String(),
			Metadata: map[string]any{
				"page": page.PageNumber,
			},
		})
	}

	return docs
}
