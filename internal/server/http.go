package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	transportconnect "github.com/year-recap/internal/transport/connect"
)

type Options struct {
	StaticDir      string
	FrontendDir    string
	AllowedOrigins []string
}

func NewHandler(application transportconnect.Application, options Options) (http.Handler, error) {
	recapHandler, err := transportconnect.NewHandler(application)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	path, handler := recapv1connect.NewRecapServiceHandler(recapHandler)

	// Keep the original Connect endpoint for local/backend-only clients.
	mux.Handle(path, handler)

	// Production frontend uses /api as its base URL. Strip that prefix before
	// handing the request to the generated Connect handler.
	mux.Handle("/api"+path, http.StripPrefix("/api", handler))
	// Explainability endpoint exposes a privacy-safe decision trace for demos and rule debugging.
	mux.HandleFunc("/explain", explainRecap(application))
	mux.HandleFunc("/api/explain", explainRecap(application))

	// Do not let unknown API URLs fall through to the React SPA.
	mux.Handle("/api/", http.NotFoundHandler())

	mux.HandleFunc("/health", health)

	if strings.TrimSpace(options.StaticDir) != "" {
		avatarsDir := filepath.Join(options.StaticDir, "avatars")
		files := http.StripPrefix("/avatars/", http.FileServer(http.Dir(avatarsDir)))
		mux.Handle("/avatars/", cacheStatic(files))
	}

	if strings.TrimSpace(options.FrontendDir) != "" {
		frontend, err := spa(options.FrontendDir)
		if err != nil {
			return nil, err
		}
		mux.Handle("/", frontend)
	}

	return cors(mux, options.AllowedOrigins), nil
}

func health(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		response.Header().Set("Allow", "GET, HEAD")
		http.Error(response, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)
	if request.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(response).Encode(map[string]string{"status": "ok"})
}

func spa(frontendDir string) (http.Handler, error) {
	root := filepath.Clean(frontendDir)
	indexPath := filepath.Join(root, "index.html")
	info, err := os.Stat(indexPath)
	if err != nil {
		return nil, fmt.Errorf("frontend index %q: %w", indexPath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("frontend index %q is a directory", indexPath)
	}

	files := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			http.NotFound(response, request)
			return
		}

		cleanURLPath := path.Clean("/" + request.URL.Path)
		relativePath := strings.TrimPrefix(cleanURLPath, "/")
		candidate := filepath.Join(root, filepath.FromSlash(relativePath))
		if candidateInfo, err := os.Stat(candidate); err == nil && !candidateInfo.IsDir() {
			files.ServeHTTP(response, request)
			return
		}

		// React Router handles client-side routes. Returning index.html here makes
		// direct links such as /recap/active-buyer work after a page refresh.
		http.ServeFile(response, request, indexPath)
	}), nil
}

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		next.ServeHTTP(response, request)
	})
}

func cors(next http.Handler, origins []string) http.Handler {
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		if value := strings.TrimSpace(origin); value != "" {
			allowed[value] = struct{}{}
		}
	}
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" || isSameOrigin(origin, request) {
			next.ServeHTTP(response, request)
			return
		}
		if _, exists := allowed[origin]; !exists {
			http.Error(response, "origin is not allowed", http.StatusForbidden)
			return
		}
		header := response.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Set("Access-Control-Expose-Headers", "Grpc-Message, Grpc-Status, Grpc-Status-Details-Bin")
		header.Add("Vary", "Origin")
		if request.Method != http.MethodOptions {
			next.ServeHTTP(response, request)
			return
		}
		if request.Header.Get("Access-Control-Request-Method") == "" {
			http.Error(response, "missing access control request method", http.StatusBadRequest)
			return
		}
		header.Set("Access-Control-Allow-Methods", "GET, HEAD, POST, OPTIONS")
		header.Set(
			"Access-Control-Allow-Headers",
			"Authorization, Connect-Protocol-Version, Connect-Timeout-Ms, Content-Type, Grpc-Timeout, X-Grpc-Web, X-User-Agent",
		)
		header.Set("Access-Control-Max-Age", "7200")
		header.Add("Vary", "Access-Control-Request-Method")
		header.Add("Vary", "Access-Control-Request-Headers")
		response.WriteHeader(http.StatusNoContent)
	})
}

func isSameOrigin(origin string, request *http.Request) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Host, request.Host)
}
