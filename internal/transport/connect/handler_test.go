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

	receivedProfileID uuid.UUID
	receivedYear      uint32
	receivedRecapID   uuid.UUID
	receivedShareID   uuid.UUID
}

func (f *fakeApplication) ListProfiles(context.Context) ([]model.Profile, error) {
	return f.profiles, f.err
}

func (f *fakeApplication) Generate(
	_ context.Context,
	profileID uuid.UUID,
	year uint32,
) (model.Recap, error) {
	f.receivedProfileID = profileID
	f.receivedYear = year
	return f.recap, f.err
}

func (f *fakeApplication) Get(_ context.Context, recapID uuid.UUID) (model.Recap, error) {
	f.receivedRecapID = recapID
	return f.recap, f.err
}

func (f *fakeApplication) GetShareCard(
	_ context.Context,
	shareID uuid.UUID,
) (model.ShareCard, error) {
	f.receivedShareID = shareID
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
	if len(profiles.Msg.Profiles) != 1 || profiles.Msg.Profiles[0].ProfileCode != testkit.Profile().Code {
		t.Fatalf("unexpected profiles: %+v", profiles.Msg.Profiles)
	}

	generated, err := handler.GenerateRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
			ProfileCode: testkit.Profile().Code,
			Year:        2025,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if generated.Msg.Recap.Id != recap.ID.String() ||
		application.receivedProfileID != testkit.ProfileID ||
		application.receivedYear != 2025 {
		t.Fatalf("unexpected generation result: %+v", generated.Msg.Recap)
	}

	got, err := handler.GetRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetRecapRequest{
			ProfileCode: testkit.Profile().Code,
			Year:        2025,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Msg.Recap.Id != recap.ID.String() || application.receivedProfileID != testkit.ProfileID {
		t.Fatalf("unexpected recap result: %+v", got.Msg.Recap)
	}

	shared, err := handler.GetPublicShare(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetPublicShareRequest{ShareId: recap.ShareID.String()}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if shared.Msg.Share.ShareId != recap.ShareID.String() ||
		application.receivedShareID != recap.ShareID {
		t.Fatalf("unexpected share result: %+v", shared.Msg.Share)
	}
}

func TestHandlerRejectsUnknownProfileCode(t *testing.T) {
	application := &fakeApplication{profiles: []model.Profile{testkit.Profile()}}
	handler, err := NewHandler(application)
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GenerateRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{ProfileCode: "does-not-exist", Year: 2025}),
	)
	if !errors.Is(err, ErrProfileCodeNotFound) {
		t.Fatalf("error = %v, want profile code not found", err)
	}
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeNotFound {
		t.Fatalf("code = %s, want not_found: %v", code, err)
	}
}

func TestHandlerRejectsNonCanonicalUUIDBeforeCallingApplication(t *testing.T) {
	application := &fakeApplication{}
	handler, err := NewHandler(application)
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GetPublicShare(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetPublicShareRequest{
			ShareId: strings.ToUpper(testkit.ShareID.String()),
		}),
	)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeInvalidArgument {
		t.Fatalf("code = %s, want invalid_argument: %v", code, err)
	}
	if application.receivedShareID != uuid.Nil {
		t.Fatalf("application called with %s", application.receivedShareID)
	}
}

func TestHandlerMapsApplicationError(t *testing.T) {
	handler, err := NewHandler(&fakeApplication{
		profiles: []model.Profile{testkit.Profile()},
		err:      application.ErrRecapNotFound,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GetRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GetRecapRequest{
			ProfileCode: testkit.Profile().Code,
			Year:        2025,
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
	handler, err := NewHandler(&fakeApplication{
		profiles: []model.Profile{testkit.Profile()},
		err:      application.ErrYearNotComplete,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = handler.GenerateRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
			ProfileCode: testkit.Profile().Code,
			Year:        uint32(testkit.Clock().Year()),
		}),
	)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeFailedPrecondition {
		t.Fatalf("code = %s, want failed_precondition: %v", code, err)
	}
}
