// Репозиторий для уже посчитанных карточек
package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/year-recap/internal/recap"
)

// Записывает карточку + следующее необходимое действие (новый уровень)
func (r *Repo) SaveRecap(ctx context.Context, rec recap.Recap) error {
	cardsBatch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO recap_cards
			(user_id, year, card_id, type, position, visibility, title, subtitle, emoji,
			 metric_value, metric_unit, metric_label, reason, cta_type, cta_label, cta_url, generated_at)
	`)
	if err != nil {
		return fmt.Errorf("prepare recap cards batch: %w", err)
	}

	for _, c := range rec.Cards {
		var metricValue *float64
		var metricUnit, metricLabel string
		if c.Metric != nil {
			v := c.Metric.Value
			metricValue = &v
			metricUnit = c.Metric.Unit
			metricLabel = c.Metric.Label
		}

		var ctaType, ctaLabel, ctaURL string
		if c.CTA != nil {
			ctaType = c.CTA.Type
			ctaLabel = c.CTA.Label
			ctaURL = c.CTA.TargetURL
		}

		if err := cardsBatch.Append(
			uint64(rec.UserID), uint16(rec.Year), c.ID, string(c.Type), uint8(c.Order), string(c.Visibility),
			c.Title, c.Subtitle, c.Emoji, metricValue, metricUnit, metricLabel, c.Reason,
			ctaType, ctaLabel, ctaURL, rec.GeneratedAt,
		); err != nil {
			return fmt.Errorf("append recap card: %w", err)
		}
	}

	if err := cardsBatch.Send(); err != nil {
		return fmt.Errorf("send recap cards batch: %w", err)
	}

	if err := r.conn.Exec(ctx, `
		INSERT INTO recap_summary
			(user_id, year, persona_code, persona_title, persona_description, persona_reason,
			 next_action_type, next_action_label, next_action_url, generated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		uint64(rec.UserID), uint16(rec.Year),
		rec.Persona.Code, rec.Persona.Title, rec.Persona.Description, rec.Persona.Reason,
		rec.NextAction.Type, rec.NextAction.Label, rec.NextAction.TargetURL,
		rec.GeneratedAt,
	); err != nil {
		return fmt.Errorf("insert recap summary: %w", err)
	}

	return nil
}

// Получаем сохраненные карточки (не идем в таблицу с сырыми событиями)
func (r *Repo) GetCachedRecap(ctx context.Context, userID recap.UserID, year int) (recap.Recap, bool, error) {
	summaryRows, err := r.conn.Query(ctx, `
		SELECT persona_code, persona_title, persona_description, persona_reason,
		       next_action_type, next_action_label, next_action_url, generated_at
		FROM recap_summary FINAL
		WHERE user_id = ? AND year = ?
		LIMIT 1
	`, uint64(userID), uint16(year))
	if err != nil {
		return recap.Recap{}, false, fmt.Errorf("query cached summary: %w", err)
	}
	defer summaryRows.Close()

	if !summaryRows.Next() {
		return recap.Recap{}, false, summaryRows.Err()
	}

	var (
		persona     recap.Persona
		nextAction  recap.NextAction
		generatedAt time.Time
	)
	if err := summaryRows.Scan(
		&persona.Code, &persona.Title, &persona.Description, &persona.Reason,
		&nextAction.Type, &nextAction.Label, &nextAction.TargetURL,
		&generatedAt,
	); err != nil {
		return recap.Recap{}, false, fmt.Errorf("scan cached summary: %w", err)
	}
	summaryRows.Close()

	cardRows, err := r.conn.Query(ctx, `
		SELECT card_id, type, position, visibility, title, subtitle, emoji,
		       metric_value, metric_unit, metric_label, reason, cta_type, cta_label, cta_url
		FROM recap_cards FINAL
		WHERE user_id = ? AND year = ?
		ORDER BY position
	`, uint64(userID), uint16(year))
	if err != nil {
		return recap.Recap{}, false, fmt.Errorf("query cached cards: %w", err)
	}
	defer cardRows.Close()

	var cards []recap.Card
	for cardRows.Next() {
		var (
			c                         recap.Card
			cardType, visibility      string
			position                  uint8
			metricValue               *float64
			metricUnit, metricLabel   string
			ctaType, ctaLabel, ctaURL string
		)
		if err := cardRows.Scan(
			&c.ID, &cardType, &position, &visibility, &c.Title, &c.Subtitle, &c.Emoji,
			&metricValue, &metricUnit, &metricLabel, &c.Reason, &ctaType, &ctaLabel, &ctaURL,
		); err != nil {
			return recap.Recap{}, false, fmt.Errorf("scan cached card: %w", err)
		}

		c.Type = recap.CardType(cardType)
		c.Visibility = recap.Visibility(visibility)
		c.Order = int(position)
		if metricValue != nil {
			c.Metric = &recap.Metric{Value: *metricValue, Unit: metricUnit, Label: metricLabel}
		}
		if ctaType != "" {
			c.CTA = &recap.NextAction{Type: ctaType, Label: ctaLabel, TargetURL: ctaURL}
		}
		cards = append(cards, c)
	}
	if err := cardRows.Err(); err != nil {
		return recap.Recap{}, false, fmt.Errorf("iterate cached cards: %w", err)
	}

	return recap.Recap{
		UserID:      userID,
		Year:        year,
		GeneratedAt: generatedAt,
		Persona:     persona,
		Cards:       cards,
		NextAction:  nextAction,
	}, true, nil
}

// Аналог VACUUM в postgres - по идее должно вызываться через джоб
func (r *Repo) OptimizeAggregates(ctx context.Context) error {
	return r.conn.Exec(ctx, "OPTIMIZE TABLE recap_category_month_agg FINAL")
}
