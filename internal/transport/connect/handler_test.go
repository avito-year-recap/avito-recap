package connect

import (
	"context"
	"errors"
	"strings"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/google/uuid"
	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
)

type fakeApplication struct {
	profiles []model.Profile
	recap    model.Recap
	share    model.ShareCard
	err      error

	receivedID   uuid.UUID
	receivedYear uint32
}

func (f *fakeApplication) ListProfiles(context.Context) ([]model.Profile, error) {
	return f.profiles, f.err
}

func (f *fakeApplication) Generate(
	_ context.Context,
	profileID uuid.UUID,
	year uint32,
) (model.Recap, error) {
	f.receivedID = profileID
	f.receivedYear = year
	return f.recap, f.err
}

func (f *fakeApplication) Get(_ context.Context, recapID uuid.UUID) (model.Recap, error) {
	f.receivedID = recapID
	return f.recap, f.err
}

func (f *fakeApplication) GetShareCard(
	_ context.Context,
	shareID uuid.UUID,
) (model.ShareCard, error) {
	f.receivedID = shareID
	return f.share, f.err
}

func TestNewHandlerRequiresApplication(t *testing.T) {
	if _, err := NewHandler(nil); !errors.Is(err, ErrMissingApplication) {
		t.Fatalf("error = %v, want missing application", err)
	}
}

func TestHandlerDelegatesAllOperations(t *testing.T) {
	recap := testkit.Recap()
	application := &fakeApplication{
		profiles: []model.Profile{testkit.Profile()},
		recap:    recap,
		share:    recap.Cards[len(recap.Cards)-1].Payload.(model.ShareCard),
	}
	handler, err := NewHandler(application)
	if err != nil {
		t.Fatal(err)
	}

	profiles, err := handler.ListProfiles(
		context.Background(),
		connectrpc.NewRequest(&recapv1.ListProfilesRequest{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles.Msg.Profiles) != 1 || profiles.Msg.Profiles[0].Id != testkit.ProfileID.String() {
		t.Fatalf("unexpected profiles: %+v", profiles.Msg.Profiles)
	}

	generated, err := handler.GenerateRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
			ProfileId: testkit.ProfileID.String(),
			Year:      2025,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if generated.Msg.Recap.InternalId != recap.ID.String() ||
		application.receivedID != testkit.ProfileID ||
		application.receivedYear != 2025 {
		t.Fatalf("unexpected generation result: %+v", generated.Msg.Recap)
	}

	got, err := handler.GetRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetRecapRequest{InternalRecapId: recap.ID.String()}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Msg.Recap.InternalId != recap.ID.String() || application.receivedID != recap.ID {
		t.Fatalf("unexpected recap result: %+v", got.Msg.Recap)
	}

	shared, err := handler.GetShareCard(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetShareCardRequest{ShareId: recap.ShareID.String()}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if shared.Msg.ShareCard.ShareId != recap.ShareID.String() ||
		application.receivedID != recap.ShareID {
		t.Fatalf("unexpected share result: %+v", shared.Msg.ShareCard)
	}
}

func TestHandlerRejectsNonCanonicalUUIDBeforeCallingApplication(t *testing.T) {
	application := &fakeApplication{}
	handler, err := NewHandler(application)
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GetShareCard(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetShareCardRequest{
			ShareId: strings.ToUpper(testkit.ShareID.String()),
		}),
	)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeInvalidArgument {
		t.Fatalf("code = %s, want invalid_argument: %v", code, err)
	}
	if application.receivedID != uuid.Nil {
		t.Fatalf("application called with %s", application.receivedID)
	}
}

func TestHandlerMapsApplicationError(t *testing.T) {
	handler, err := NewHandler(&fakeApplication{err: application.ErrRecapNotFound})
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GetRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetRecapRequest{
			InternalRecapId: testkit.RecapID.String(),
		}),
	)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeNotFound {
		t.Fatalf("code = %s, want not_found: %v", code, err)
	}
}

func TestHandlerRejectsNilGenerateRequest(t *testing.T) {
	handler, err := NewHandler(&fakeApplication{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GenerateRecap(context.Background(), nil)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeInvalidArgument {
		t.Fatalf("code = %s, want invalid_argument: %v", code, err)
	}
}

func TestHandlerMapsFailedPreconditionForUnfinishedYear(t *testing.T) {
	handler, err := NewHandler(&fakeApplication{err: application.ErrYearNotComplete})
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GenerateRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
			ProfileId: testkit.ProfileID.String(),
			Year:      uint32(testkit.Clock().Year()),
		}),
	)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeFailedPrecondition {
		t.Fatalf("code = %s, want failed_precondition: %v", code, err)
	}
}
