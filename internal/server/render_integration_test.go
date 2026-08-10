package server_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	connectrpc "connectrpc.com/connect"
	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/server"
	"github.com/year-recap/internal/storage/memory"
)

func TestRenderSingleServiceServesFrontendAndDeepLinks(t *testing.T) {
	httpServer := newRenderStyleServer(t)

	for _, route := range []string{"/", "/recap/active-buyer", "/share/example-share-id"} {
		response, err := httpServer.Client().Get(httpServer.URL + route)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		closeResponseBody(t, response)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", route, response.StatusCode)
		}
		if !strings.Contains(string(body), "render-integration-spa") {
			t.Fatalf("GET %s did not return SPA index: %s", route, body)
		}
	}

	response, err := httpServer.Client().Get(httpServer.URL + "/assets/app.js")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	closeResponseBody(t, response)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || string(body) != "console.log('render integration');" {
		t.Fatalf("asset response status=%d body=%q", response.StatusCode, body)
	}
}

func TestRenderSingleServiceMainRecapFlowThroughAPIPrefix(t *testing.T) {
	httpServer := newRenderStyleServer(t)
	client := recapv1connect.NewRecapServiceClient(httpServer.Client(), httpServer.URL+"/api")
	ctx := context.Background()

	profiles, err := client.ListProfiles(ctx, connectrpc.NewRequest(&recapv1.ListProfilesRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles.Msg.Profiles) != 17 {
		t.Fatalf("profile count = %d, want 17", len(profiles.Msg.Profiles))
	}

	profile := profiles.Msg.Profiles[0]
	generated, err := client.GenerateRecap(ctx, connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
		ProfileCode: profile.ProfileCode,
		Year:        2025,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if generated.Msg.Recap == nil || generated.Msg.Recap.Id == "" {
		t.Fatalf("generated recap is incomplete: %+v", generated.Msg.Recap)
	}

	fetched, err := client.GetRecap(ctx, connectrpc.NewRequest(&recapv1.GetRecapRequest{
		ProfileCode: profile.ProfileCode,
		Year:        2025,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if fetched.Msg.Recap.Id != generated.Msg.Recap.Id {
		t.Fatalf("GetRecap id = %q, want %q", fetched.Msg.Recap.Id, generated.Msg.Recap.Id)
	}

	shareCard := findCard(t, generated.Msg.Recap.Cards, recapv1.CardType_CARD_TYPE_SHARE).GetShare()
	shared, err := client.GetPublicShare(ctx, connectrpc.NewRequest(&recapv1.GetPublicShareRequest{
		ShareId: shareCard.ShareId,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if shared.Msg.Share == nil || shared.Msg.Share.ShareId != shareCard.ShareId {
		t.Fatalf("public share mismatch: %+v", shared.Msg.Share)
	}

	avatar, err := httpServer.Client().Get(httpServer.URL + profile.GetAvatarUrl())
	if err != nil {
		t.Fatal(err)
	}
	avatarBody, err := io.ReadAll(avatar.Body)
	closeResponseBody(t, avatar)
	if err != nil {
		t.Fatal(err)
	}
	if avatar.StatusCode != http.StatusOK || len(avatarBody) == 0 {
		t.Fatalf("avatar status=%d size=%d", avatar.StatusCode, len(avatarBody))
	}
}

func TestRenderSingleServiceSameOriginAndUnknownAPIBehavior(t *testing.T) {
	httpServer := newRenderStyleServer(t)

	request, err := http.NewRequest(http.MethodGet, httpServer.URL+"/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", httpServer.URL)
	response, err := httpServer.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	closeResponseBody(t, response)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("same-origin request status = %d, want 200", response.StatusCode)
	}

	response, err = httpServer.Client().Get(httpServer.URL + "/api/not-a-real-endpoint")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	closeResponseBody(t, response)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown API status = %d, want 404", response.StatusCode)
	}
	if strings.Contains(string(body), "render-integration-spa") {
		t.Fatalf("unknown API route fell through to SPA: %s", body)
	}
}

func newRenderStyleServer(t *testing.T) *httptest.Server {
	t.Helper()

	root := projectRoot(t)
	store, err := memory.Load(
		filepath.Join(root, "seeds", "profiles.json"),
		filepath.Join(root, "seeds", "scenarios.json"),
	)
	if err != nil {
		t.Fatal(err)
	}
	service, err := application.NewService(
		store,
		store,
		store,
		store,
		application.WithClock(testkit.Clock),
	)
	if err != nil {
		t.Fatal(err)
	}

	frontendDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(frontendDir, "index.html"),
		[]byte(`<!doctype html><html><body>render-integration-spa</body></html>`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(frontendDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(frontendDir, "assets", "app.js"),
		[]byte("console.log('render integration');"),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(frontendDir, "avatars"), 0o755); err != nil {
		t.Fatal(err)
	}
	avatarData, err := os.ReadFile(filepath.Join(root, "frontend", "public", "avatars", "active-buyer.png"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(frontendDir, "avatars", "active-buyer.png"), avatarData, 0o644); err != nil {
		t.Fatal(err)
	}

	handler, err := server.NewHandler(service, server.Options{
		StaticDir:   frontendDir,
		FrontendDir: frontendDir,
	})
	if err != nil {
		t.Fatal(err)
	}

	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)
	return httpServer
}
