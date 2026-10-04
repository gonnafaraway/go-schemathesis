package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
	"github.com/gonnafaraway/go-schemathesis/internal/check"
	"github.com/gonnafaraway/go-schemathesis/internal/config"
	"github.com/gonnafaraway/go-schemathesis/internal/httpx"
	"github.com/gonnafaraway/go-schemathesis/internal/report"
	"github.com/gonnafaraway/go-schemathesis/internal/schema"
)

// ErrFailuresFound indicates the run completed with check failures.
var ErrFailuresFound = fmt.Errorf("one or more checks failed")

// SchemaLoader loads an OpenAPI document.
type SchemaLoader interface {
	Load(ctx context.Context, source string) (*schema.Spec, error)
}

// Runner orchestrates schema load, generation, execution and checks.
type Runner struct {
	cfg    config.Config
	loader SchemaLoader
	client httpx.Client
	log    *slog.Logger
}

// Dependencies are optional overrides for tests.
type Dependencies struct {
	Loader SchemaLoader
	Client httpx.Client
	Log    *slog.Logger
}

// New constructs a runner from config and optional dependencies.
func New(cfg config.Config, deps Dependencies) (*Runner, error) {
	normalized, err := config.Normalize(cfg)
	if err != nil {
		return nil, err
	}

	loader := deps.Loader
	if loader == nil {
		loader = schema.NewLoader()
	}

	log := deps.Log
	if log == nil {
		opts := &slog.HandlerOptions{Level: slog.LevelInfo}
		if normalized.Verbose {
			opts.Level = slog.LevelDebug
		}
		log = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return &Runner{
		cfg:    normalized,
		loader: loader,
		client: deps.Client,
		log:    log,
	}, nil
}

// Run executes the full test pipeline and returns a summary.
func (r *Runner) Run(ctx context.Context) (*report.Summary, error) {
	started := time.Now()
	r.log.Info("loading schema", "source", r.cfg.SchemaPath)

	spec, err := r.loader.Load(ctx, r.cfg.SchemaPath)
	if err != nil {
		return nil, fmt.Errorf("load schema: %w", err)
	}
	if r.cfg.BaseURL == "" {
		r.cfg.BaseURL = spec.BaseURL
	}
	if r.cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required: pass --url or define servers in the schema")
	}
	if r.client == nil {
		r.client, err = httpx.NewClient(r.cfg.BaseURL, r.cfg.RequestTimeout, r.cfg.Headers)
		if err != nil {
			return nil, err
		}
	}

	r.log.Info("schema loaded", "title", spec.Title, "operations", len(spec.Operations))

	gen := casegen.New(casegen.Options{
		Phases:      r.cfg.Phases,
		Mode:        r.cfg.Mode,
		MaxExamples: r.cfg.MaxExamples,
		Seed:        r.cfg.Seed,
	})
	cases, err := gen.Generate(spec)
	if err != nil {
		return nil, fmt.Errorf("generate cases: %w", err)
	}
	r.log.Info("generated cases", "count", len(cases))

	registry, err := check.NewRegistry(r.cfg.Checks)
	if err != nil {
		return nil, err
	}

	opsByID := indexOperations(spec)
	summary := &report.Summary{
		BaseURL:    r.cfg.BaseURL,
		StartedAt:  started,
		TotalCases: len(cases),
		Failures:   make([]check.Failure, 0),
	}

	var (
		mu          sync.Mutex
		passed      atomic.Int64
		failed      atomic.Int64
		errorsCount atomic.Int64
		stop        atomic.Bool
	)

	jobs := make(chan casegen.Case)
	var wg sync.WaitGroup
	workers := r.cfg.Workers
	if workers < 1 {
		workers = 1
	}

	workerFn := func() {
		defer wg.Done()
		for c := range jobs {
			if stop.Load() {
				continue
			}
			if err := ctx.Err(); err != nil {
				return
			}
			result, execErr := r.executeOne(ctx, c, opsByID[c.OperationID], registry)
			if execErr != nil {
				errorsCount.Add(1)
				r.log.Error("case execution failed", "operation", c.OperationID, "err", execErr)
				continue
			}
			if result.Passed {
				passed.Add(1)
				continue
			}
			failed.Add(1)
			mu.Lock()
			for i := range result.Failures {
				result.Failures[i].CurlCommand = report.Curl(r.cfg.BaseURL, result.Failures[i].Case)
				summary.Failures = append(summary.Failures, result.Failures[i])
			}
			shouldStop := r.cfg.MaxFailures > 0 && int(failed.Load()) >= r.cfg.MaxFailures
			mu.Unlock()
			if shouldStop {
				stop.Store(true)
			}
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go workerFn()
	}
	for _, c := range cases {
		if stop.Load() {
			break
		}
		jobs <- c
	}
	close(jobs)
	wg.Wait()

	summary.PassedCases = int(passed.Load())
	summary.FailedCases = int(failed.Load())
	summary.Errors = int(errorsCount.Load())
	summary.FinishedAt = time.Now()
	return summary, nil
}

func (r *Runner) executeOne(
	ctx context.Context,
	c casegen.Case,
	op schema.Operation,
	registry *check.Registry,
) (check.Result, error) {
	req := httpx.FromCase(r.cfg.BaseURL, c)
	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return check.Result{}, err
	}

	return registry.Evaluate(check.Context{
		Case:      c,
		Operation: op,
		Response: check.Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Headers,
			Body:       resp.Body,
			ElapsedMS:  resp.ElapsedMS,
		},
	}), nil
}

func indexOperations(spec *schema.Spec) map[string]schema.Operation {
	out := make(map[string]schema.Operation, len(spec.Operations))
	for _, op := range spec.Operations {
		out[op.ID] = op
	}
	return out
}
