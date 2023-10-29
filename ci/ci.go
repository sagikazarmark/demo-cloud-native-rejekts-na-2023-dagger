package main

import (
	"context"

	"golang.org/x/sync/errgroup"
)

const (
	goVersion           = "1.21.3"
	golangciLintVersion = "1.54.2"
)

type Ci struct{}

// Run the entire CI pipeline
func (m *Ci) CI(ctx context.Context) error {
	var group errgroup.Group

	// Build
	{
		dag := dag.Pipeline("Build")
		_ = dag
	}

	// Test
	{
		dag := dag.Pipeline("Test")

		group.Go(func() error {
			return testAll(ctx, dag)
		})
	}

	// Lint
	{
		dag := dag.Pipeline("Lint")

		group.Go(func() error {
			return lintAll(ctx, dag)
		})
	}

	return group.Wait()
}

// Build jobs
func (m *Ci) Build() *Build {
	return &Build{}
}

// Test jobs
func (m *Ci) Test() *Test {
	return &Test{}
}

// Linter jobs
func (m *Ci) Lint() *Lint {
	return &Lint{}
}
