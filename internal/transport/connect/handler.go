package connect

import (
	"context"
	"errors"
	"fmt"

	connectrpc "connectrpc.com/connect"
	"github.com/google/uuid"
	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
)

var ErrMissingApplication = errors.New("missing recap application")

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
) (*connectrpc.Response[recapv1.GenerateRecapResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, invalidArgumentError("request is required")
	}
	profileID, err := parseCanonicalUUID("profile_id", request.Msg.ProfileId)
	if err != nil {
		return nil, err
	}
	value, err := h.application.Generate(ctx, profileID, request.Msg.Year)
	if err != nil {
		return nil, transportError(err)
	}
	mapped, err := recapToProto(value)
	if err != nil {
		return nil, transportError(err)
	}
	return connectrpc.NewResponse(&recapv1.GenerateRecapResponse{Recap: mapped}), nil
}

func (h *Handler) GetRecap(
	ctx context.Context,
	request *connectrpc.Request[recapv1.GetRecapRequest],
) (*connectrpc.Response[recapv1.GetRecapResponse], error) {
	if request == nil || request.Msg == nil {
		return nil, invalidArgumentError("request is required")
	}
	recapID, err := parseCanonicalUUID("internal_recap_id", request.Msg.InternalRecapId)
	if err != nil {
		return nil, err
	}
	value, err := h.application.Get(ctx, recapID)
	if err != nil {
		return nil, transportError(err)
	}
	mapped, err := recapToProto(value)
	if err != nil {
		return nil, transportError(err)
	}
	return connectrpc.NewResponse(&recapv1.GetRecapResponse{Recap: mapped}), nil
}

func (h *Handler) GetShareCard(
	ctx context.Context,
	request *connectrpc.Request[recapv1.GetShareCardRequest],
) (*connectrpc.Response[recapv1.GetShareCardResponse], error) {
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
	return connectrpc.NewResponse(&recapv1.GetShareCardResponse{
		ShareCard: shareCardToProto(value),
	}), nil
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
