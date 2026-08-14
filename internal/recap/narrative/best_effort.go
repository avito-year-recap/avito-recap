package narrative

import (
	"context"
	"fmt"

	"github.com/year-recap/internal/recap/model"
)

// BestEffort owns the complete optional-AI enrichment step. Provider failures
// and invalid AI copy are reported through the same callback and fall back to
// the deterministic recap. Request cancellation is still propagated.
type BestEffort struct {
	Primary Generator
	OnError func(error)
}

func (g BestEffort) Enrich(ctx context.Context, recap model.Recap) (model.Recap, error) {
	if g.Primary == nil {
		return recap, nil
	}

	story, err := g.Primary.Generate(ctx, FactsFromRecap(recap))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return recap, ctxErr
		}
		g.report(fmt.Errorf("generate AI narrative: %w", err))
		return recap, nil
	}
	if err := ctx.Err(); err != nil {
		return recap, err
	}

	enriched, err := Apply(recap, story)
	if err != nil {
		g.report(fmt.Errorf("apply AI narrative: %w", err))
		return recap, nil
	}
	return enriched, nil
}

func (g BestEffort) report(err error) {
	if g.OnError != nil {
		g.OnError(err)
	}
}
