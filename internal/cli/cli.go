package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gonnafaraway/go-schemathesis/internal/casegen"
	"github.com/gonnafaraway/go-schemathesis/internal/check"
	"github.com/gonnafaraway/go-schemathesis/internal/config"
	"github.com/gonnafaraway/go-schemathesis/internal/report"
	"github.com/gonnafaraway/go-schemathesis/internal/runner"
)

// Version is set at build time when needed.
var Version = "0.1.0"

// Execute is the CLI composition root.
func Execute(args []string) error {
	root := newRootCommand()
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return nil
	}
	if errors.Is(err, runner.ErrFailuresFound) {
		os.Exit(1)
	}
	return err
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "schemathesis",
		Short:         "Property-based API testing from OpenAPI schemas",
		Long:          "Go port of Schemathesis-style OpenAPI testing: generate cases, execute requests, validate responses.",
		Version:       Version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.AddCommand(newRunCommand())
	return root
}

func newRunCommand() *cobra.Command {
	var (
		baseURL        string
		workers        string
		phases         string
		mode           string
		checks         string
		excludeChecks  string
		maxExamples    int
		maxFailures    int
		requestTimeout time.Duration
		headerFlags    []string
		seed           int64
		verbose        bool
	)

	cmd := &cobra.Command{
		Use:   "run SCHEMA",
		Short: "Run API tests generated from an OpenAPI schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workerCount, err := parseWorkers(workers)
			if err != nil {
				return err
			}
			parsedPhases, err := parsePhases(phases)
			if err != nil {
				return err
			}
			parsedChecks, err := parseChecks(checks)
			if err != nil {
				return err
			}
			parsedExclude, err := parseChecks(excludeChecks)
			if err != nil {
				return err
			}
			headers, err := parseHeaders(headerFlags)
			if err != nil {
				return err
			}

			cfg := config.Config{
				SchemaPath:     args[0],
				BaseURL:        baseURL,
				Workers:        workerCount,
				Phases:         parsedPhases,
				Mode:           casegen.Mode(mode),
				Checks:         parsedChecks,
				ExcludeChecks:  parsedExclude,
				MaxExamples:    maxExamples,
				MaxFailures:    maxFailures,
				RequestTimeout: requestTimeout,
				Headers:        headers,
				Seed:           seed,
				Verbose:        verbose,
			}
			return runCommand(cmd.Context(), cfg)
		},
	}

	cmd.Flags().StringVarP(&baseURL, "url", "u", "", "Base URL of the API under test")
	cmd.Flags().StringVarP(&workers, "workers", "w", "1", "Number of concurrent workers or \"auto\"")
	cmd.Flags().StringVar(&phases, "phases", "examples,coverage,fuzzing", "Comma-separated phases")
	cmd.Flags().StringVar(&mode, "mode", "all", "Generation mode: all|positive|negative")
	cmd.Flags().StringVarP(&checks, "checks", "c", "all", "Comma-separated checks or \"all\"")
	cmd.Flags().StringVar(&excludeChecks, "exclude-checks", "", "Comma-separated checks to skip")
	cmd.Flags().IntVar(&maxExamples, "max-examples", 100, "Max fuzzing examples per operation")
	cmd.Flags().IntVar(&maxFailures, "max-failures", 0, "Stop after N failures (0 = disabled)")
	cmd.Flags().DurationVar(&requestTimeout, "request-timeout", 10*time.Second, "HTTP request timeout")
	cmd.Flags().StringArrayVarP(&headerFlags, "header", "H", nil, "Extra request header (Name: Value)")
	cmd.Flags().Int64Var(&seed, "seed", 0, "RNG seed (0 = current time)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose logging")

	return cmd
}

func runCommand(ctx context.Context, cfg config.Config) error {
	r, err := runner.New(cfg, runner.Dependencies{})
	if err != nil {
		return err
	}

	summary, err := r.Run(ctx)
	if err != nil {
		return err
	}

	if err := (report.Console{Out: os.Stdout}).WriteSummary(*summary, summary.BaseURL); err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	if summary.HasFailures() {
		return runner.ErrFailuresFound
	}
	return nil
}

func parseWorkers(value string) (int, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "auto" {
		return -1, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid workers value %q", value)
	}
	if n < 1 || n > 64 {
		return 0, fmt.Errorf("workers must be between 1 and 64")
	}
	return n, nil
}

func parsePhases(value string) ([]casegen.Phase, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parts := splitCSV(value)
	out := make([]casegen.Phase, 0, len(parts))
	for _, part := range parts {
		phase := casegen.Phase(part)
		switch phase {
		case casegen.PhaseExamples, casegen.PhaseCoverage, casegen.PhaseFuzzing:
			out = append(out, phase)
		default:
			return nil, fmt.Errorf("unsupported phase %q", part)
		}
	}
	return out, nil
}

func parseChecks(value string) ([]check.Name, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	if value == "all" {
		return check.AllNames(), nil
	}
	parts := splitCSV(value)
	out := make([]check.Name, 0, len(parts))
	known := map[check.Name]struct{}{}
	for _, name := range check.AllNames() {
		known[name] = struct{}{}
	}
	for _, part := range parts {
		name := check.Name(part)
		if _, ok := known[name]; !ok {
			return nil, fmt.Errorf("unknown check %q", part)
		}
		out = append(out, name)
	}
	return out, nil
}

func parseHeaders(values []string) (map[string]string, error) {
	out := map[string]string{}
	for _, raw := range values {
		name, value, ok := strings.Cut(raw, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q, expected Name: Value", raw)
		}
		out[strings.TrimSpace(name)] = strings.TrimSpace(value)
	}
	return out, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}
