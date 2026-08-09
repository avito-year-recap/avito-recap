package connect

import (
	"context"
	"errors"
	"fmt"

	connectrpc "connectrpc.com/connect"
	"github.com/google/uuid"
	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	"github.com/year-recap/internal/recap/model"
)

var (
	ErrMissingApplication  = errors.New("missing recap application")
	ErrProfileCodeNotFound = errors.New("profile code not found")
)

type Handler struct {
	application Application
}

var _ recapv1connect.RecapServiceHandler = (*Handler)(nil)

func NewHandler(application Application) (*Handler, error) {
	if application == nil {
		return nil, ErrMissingApplication
	}
	return &Handler{application: application}, nil
}

func (h *Handler) ListProfiles(
	ctx context.Context,
	_ *connectrpc.Request[recapv1.ListProfilesRequest],
) (*connectrpc.Response[recapv1.ListProfilesResponse], error) {
	profiles, err := h.application.ListProfiles(ctx)
	if err != nil {
		return nil, transportError(err)
	}
	mapped := make([]*recapv1.Profile, 0, len(profiles))
	for _, profile := range profiles {
		mapped = append(mapped, profileToProto(profile))
	}
	return connectrpc.NewResponse(&recapv1.ListProfilesResponse{Profiles: mapped}), nil
}

func (h *Handler) GenerateRecap(
	ctx context.Context,
	request *connectrpc.Request[recapv1.GenerateRecapRequest],
) (*connectrpc.Response[recapv1.RecapResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, invalidArgumentError("request is required")
	}
	profile, err := h.resolveProfileByCode(ctx, request.Msg.ProfileCode)
	if err != nil {
		return nil, err
	}
	value, err := h.application.Generate(ctx, profile.ID, request.Msg.Year)
	if err != nil {
		return nil, transportError(err)
	}
	return recapResponse(value)
}

// GetRecap intentionally shares Generate's idempotent get-or-create path:
// the wire contract addresses a recap by (profile_code, year), and the
// application layer already guarantees that generating twice for the same
// key returns the previously stored recap rather than creating a second one.
// A read-only lookup by (profile, year) belongs to the service layer and can
// replace this once it exists.
func (h *Handler) GetRecap(
	ctx context.Context,
	request *connectrpc.Request[recapv1.GetRecapRequest],
) (*connectrpc.Response[recapv1.RecapResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, invalidArgumentError("request is required")
	}
	profile, err := h.resolveProfileByCode(ctx, request.Msg.ProfileCode)
	if err != nil {
		return nil, err
	}
	value, err := h.application.Generate(ctx, profile.ID, request.Msg.Year)
	if err != nil {
		return nil, transportError(err)
	}
	return recapResponse(value)
}

func (h *Handler) GetPublicShare(
	ctx context.Context,
	request *connectrpc.Request[recapv1.GetPublicShareRequest],
) (*connectrpc.Response[recapv1.GetPublicShareResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, invalidArgumentError("request is required")
	}
	shareID, err := parseCanonicalUUID("share_id", request.Msg.ShareId)
	if err != nil {
		return nil, err
	}
	value, err := h.application.GetShareCard(ctx, shareID)
	if err != nil {
		return nil, transportError(err)
	}
	return connectrpc.NewResponse(&recapv1.GetPublicShareResponse{
		Share: publicShareToProto(value),
	}), nil
}

func recapResponse(value model.Recap) (*connectrpc.Response[recapv1.RecapResponse], error) {
	profile := profileToProto(value.Profile)
	recap, err := recapToProto(value)
	if err != nil {
		return nil, transportError(err)
	}
	return connectrpc.NewResponse(&recapv1.RecapResponse{Profile: profile, Recap: recap}), nil
}

// resolveProfileByCode is transport-layer glue: the wire contract addresses
// profiles by their human-readable code, but the application/storage layer
// only knows how to look profiles up by internal UUID. Until the service
// layer grows a ProfileStorage.GetProfileByCode port, this scans
// ListProfiles, which is fine at demo/seed scale but not the long-term shape.
func (h *Handler) resolveProfileByCode(ctx context.Context, code string) (model.Profile, error) {
	if code == "" {
		return model.Profile{}, invalidArgumentError("profile_code is required")
	}
	profiles, err := h.application.ListProfiles(ctx)
	if err != nil {
		return model.Profile{}, transportError(err)
	}
	for _, profile := range profiles {
		if profile.Code == code {
			return profile, nil
		}
	}
	return model.Profile{}, connectrpc.NewError(
		connectrpc.CodeNotFound,
		fmt.Errorf("%w: %q", ErrProfileCodeNotFound, code),
	)
}

func parseCanonicalUUID(field, value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		if err == nil {
			err = errors.New("must use the lowercase canonical UUID form")
		}
		return uuid.Nil, connectrpc.NewError(
			connectrpc.CodeInvalidArgument,
			fmt.Errorf("%s: %w", field, err),
		)
	}
	return parsed, nil
}

func invalidArgumentError(message string) error {
	return connectrpc.NewError(connectrpc.CodeInvalidArgument, errors.New(message))
}
