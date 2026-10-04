package runner_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
	"github.com/gonnafaraway/go-schemathesis/internal/config"
	"github.com/gonnafaraway/go-schemathesis/internal/runner"
)

func TestRunAgainstMockAPI(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/pets":
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "Rex"}})
		case r.Method == http.MethodPost && r.URL.Path == "/pets":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Rex"})
		case r.Method == http.MethodGet && r.URL.Path == "/pets/1":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 1, "name": "Rex"})
		default:
			if r.Method == http.MethodGet && len(r.URL.Path) > len("/pets/") {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"message":"not found"}`))
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"bad request"}`))
		}
	}))
	t.Cleanup(server.Close)

	r, err := runner.New(config.Config{
		SchemaPath:  filepath.Join("..", "..", "examples", "petstore.yaml"),
		BaseURL:     server.URL,
		Workers:     2,
		Phases:      []casegen.Phase{casegen.PhaseExamples, casegen.PhaseCoverage},
		Mode:        casegen.ModePositive,
		MaxExamples: 5,
		Seed:        42,
	}, runner.Dependencies{
		Log: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("new runner: %v", err)
	}

	summary, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if summary.TotalCases == 0 {
		t.Fatal("expected generated cases")
	}
	if summary.Errors > 0 {
		t.Fatalf("unexpected execution errors: %d", summary.Errors)
	}
}
