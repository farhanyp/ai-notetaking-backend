package chunking

import (
	"context"
	"fmt"
	"io"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

func SplitText(ctx context.Context, input io.Reader, chunkSize int, overlapPercent int) ([]schema.Document, error) {
	loader := documentloaders.NewText(input)

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
