package chunking

import (
	"context"
	"fmt"
	"io"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// SplitText memproses data dengan persentase overlap dinamis
func SplitText(ctx context.Context, input io.Reader, chunkSize int, overlapPercent int) ([]schema.Document, error) {
	loader := documentloaders.NewText(input)

	// Hitung overlap berdasarkan persentase dari chunkSize
	// Contoh: chunkSize 1000, overlapPercent 10 -> overlap = 100 karakter
	calculatedOverlap := (chunkSize * overlapPercent) / 100

	split := textsplitter.NewRecursiveCharacter()
	split.ChunkSize = chunkSize
	split.ChunkOverlap = calculatedOverlap

	docs, err := loader.LoadAndSplit(ctx, split)
	if err != nil {
		return nil, fmt.Errorf("split error: %w", err)
	}

	return docs, nil
}
