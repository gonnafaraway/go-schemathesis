package config

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
	"github.com/gonnafaraway/go-schemathesis/internal/check"
)

// Config holds runtime configuration for a schemathesis run.
type Config struct {
	SchemaPath     string
	BaseURL        string
	Workers        int
	Phases         []casegen.Phase
	Mode           casegen.Mode
	Checks         []check.Name
	ExcludeChecks  []check.Name
	MaxExamples    int
	MaxFailures    int
	RequestTimeout time.Duration
	Headers        map[string]string
	Seed           int64
	Verbose        bool
}

// Normalize applies defaults and validates configuration.
func Normalize(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.SchemaPath) == "" {
		return Config{}, fmt.Errorf("schema path is required")
	}

	if cfg.Workers == -1 {
		cfg.Workers = runtime.NumCPU()
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	if cfg.Workers > 64 {
		cfg.Workers = 64
	}

	if len(cfg.Phases) == 0 {
		cfg.Phases = []casegen.Phase{
			casegen.PhaseExamples,
			casegen.PhaseCoverage,
			casegen.PhaseFuzzing,
		}
	}

	if cfg.Mode == "" {
		cfg.Mode = casegen.ModeAll
	}
	switch cfg.Mode {
	case casegen.ModePositive, casegen.ModeNegative, casegen.ModeAll:
	default:
		return Config{}, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}

	if len(cfg.Checks) == 0 {
		cfg.Checks = check.AllNames()
	}
	cfg.Checks = filterChecks(cfg.Checks, cfg.ExcludeChecks)

	if cfg.MaxExamples <= 0 {
		cfg.MaxExamples = 100
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.Headers == nil {
		cfg.Headers = map[string]string{}
	}
	if cfg.Seed == 0 {
		cfg.Seed = time.Now().UnixNano()
	}

	return cfg, nil
}

func filterChecks(enabled, excluded []check.Name) []check.Name {
	if len(excluded) == 0 {
		return enabled
	}
	deny := make(map[check.Name]struct{}, len(excluded))
	for _, name := range excluded {
		deny[name] = struct{}{}
	}
	out := make([]check.Name, 0, len(enabled))
	for _, name := range enabled {
		if _, skip := deny[name]; skip {
			continue
		}
		out = append(out, name)
	}
	return out
}
