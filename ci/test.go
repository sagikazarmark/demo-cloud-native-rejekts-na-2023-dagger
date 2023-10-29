package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type Test struct{}

// Run all tests
func (m *Test) All(ctx context.Context) error {
	return testAll(ctx, dag)
}

func testAll(ctx context.Context, dag *Client) error {
	var group errgroup.Group

	group.Go(func() error {
		_, err := testGo(dag).Sync(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	return group.Wait()
}

// Run Go tests
func (m *Test) Go() *Container {
	return testGo(dag)
}

func testGo(dag *Client) *Container {
	host := dag.Host()

	return dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedDirectory("/src", host.Directory(root(), HostDirectoryOpts{
			Exclude: []string{".direnv", ".devenv", "api/client/node/node_modules"},
		})).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-v", "./..."})
}
