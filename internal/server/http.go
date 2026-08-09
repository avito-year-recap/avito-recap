package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	transportconnect "github.com/year-recap/internal/transport/connect"
)

type Options struct {
	StaticDir      string
	AllowedOrigins []string
}

func NewHandler(application transportconnect.Application, options Options) (http.Handler, error) {
	recapHandler, err := transportconnect.NewHandler(application)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	path, handler := recapv1connect.NewRecapServiceHandler(recapHandler)
	mux.Handle(path, handler)
	mux.HandleFunc("/health", health)
	if strings.TrimSpace(options.StaticDir) != "" {
		avatarsDir := filepath.Join(options.StaticDir, "avatars")
		files := http.StripPrefix("/avatars/", http.FileServer(http.Dir(avatarsDir)))
		mux.Handle("/avatars/", cacheStatic(files))
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

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Cache-Control", "public, max-age=3600")
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
		if origin == "" {
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
