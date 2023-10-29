package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Lint struct{}

// Run all linters
func (m *Lint) All(ctx context.Context) error {
	return lintAll(ctx, dag)
}

func lintAll(ctx context.Context, dag *Client) error {
	var group errgroup.Group

	group.Go(func() error {
		_, err := lintGo(dag).Sync(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	return group.Wait()
}

// Run Go linters
func (m *Lint) Go() *Container {
	return lintGo(dag)
}

func lintGo(dag *Client) *Container {
	host := dag.Host()

	bin := dag.Container().
		From(fmt.Sprintf("docker.io/golangci/golangci-lint:v%s", golangciLintVersion)).
		File("/usr/bin/golangci-lint")

	return dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedDirectory("/src", host.Directory(root(), HostDirectoryOpts{
			Exclude: []string{".direnv", ".devenv", "api/client/node/node_modules"},
		})).
		WithWorkdir("/src").
		WithFile("/usr/local/bin/golangci-lint", bin).
		WithExec([]string{"golangci-lint", "run", "--verbose"})
}
