package clickhouse

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
)

// Каталог тестовых профилей
func (r *Repo) ListProfiles(ctx context.Context) ([]model.Profile, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, code, display_name, description, avatar_url
		FROM profiles FINAL
		ORDER BY code
	`)
	if err != nil {
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close profiles rows: %v", err)
		}
	}()

	var profiles []model.Profile
	for rows.Next() {
		var profile model.Profile
		if err := rows.Scan(&profile.ID, &profile.Code, &profile.DisplayName, &profile.Description, &profile.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

// Профиль пользователя
func (r *Repo) GetProfile(ctx context.Context, profileID uuid.UUID) (model.Profile, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, code, display_name, description, avatar_url
		FROM profiles FINAL
		WHERE id = ?
		LIMIT 1
	`, profileID)
	if err != nil {
		return model.Profile{}, fmt.Errorf("query profile: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close profile rows: %v", err)
		}
	}()

	if !rows.Next() {
		return model.Profile{}, application.ErrRecapNotFound
	}
	var profile model.Profile
	if err := rows.Scan(&profile.ID, &profile.Code, &profile.DisplayName, &profile.Description, &profile.AvatarURL); err != nil {
		return model.Profile{}, fmt.Errorf("scan profile: %w", err)
	}
	return profile, rows.Err()
}

// Профиль по code
func (r *Repo) GetProfileByCode(ctx context.Context, code string) (model.Profile, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, code, display_name, description, avatar_url
		FROM profiles FINAL
		WHERE code = ?
		LIMIT 1
	`, code)
	if err != nil {
		return model.Profile{}, fmt.Errorf("query profile by code: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close profile by code rows: %v", err)
		}
	}()

	if !rows.Next() {
		return model.Profile{}, application.ErrProfileNotFound
	}
	var profile model.Profile
	if err := rows.Scan(&profile.ID, &profile.Code, &profile.DisplayName, &profile.Description, &profile.AvatarURL); err != nil {
		return model.Profile{}, fmt.Errorf("scan profile: %w", err)
	}
	return profile, rows.Err()
}

// Обновление профиля
func (r *Repo) UpsertProfiles(ctx context.Context, profiles []model.Profile) error {
	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO profiles (id, code, display_name, description, avatar_url)
	`)
	if err != nil {
		return fmt.Errorf("prepare profiles batch: %w", err)
	}
	for _, profile := range profiles {
		if err := batch.Append(profile.ID, profile.Code, profile.DisplayName, profile.Description, profile.AvatarURL); err != nil {
			return fmt.Errorf("append profile %s: %w", profile.Code, err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("send profiles batch: %w", err)
	}
	return nil
}
