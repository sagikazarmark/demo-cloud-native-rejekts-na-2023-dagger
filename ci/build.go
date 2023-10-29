package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

const (
	artifactBaseImage = "alpine:3.18.3"
)

type Build struct{}

// Run all linters
func (m *Build) All(ctx context.Context) error {
	return buildAll(ctx, dag)
}

func buildAll(ctx context.Context, dag *Client) error {
	var group errgroup.Group

	group.Go(func() error {
		_, err := buildBinary(dag).Sync(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	group.Go(func() error {
		_, err := buildContainerImage(dag).Sync(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	return group.Wait()
}

// Build the binary
func (m *Build) Binary(ctx context.Context) *File {
	return buildBinary(dag)
}

func buildBinary(dag *Client) *File {
	host := dag.Host()

	return dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedDirectory("/src", host.Directory(root(), HostDirectoryOpts{
			Exclude: []string{".direnv", ".devenv"},
		})).
		WithWorkdir("/src").
		WithExec([]string{"go", "build", "-ldflags", "-X main.version=${VERSION}", "-o", "/usr/local/bin/app", "."}).
		File("/usr/local/bin/app")
}

// Build the container image
func (m *Build) ContainerImage(ctx context.Context) *Container {
	return buildContainerImage(dag)
}

func buildContainerImage(dag *Client) *Container {
	binary := buildBinary(dag)

	return dag.Container().From(artifactBaseImage).
		WithExec([]string{"apk", "add", "--update", "--no-cache", "ca-certificates", "tzdata"}).
		WithFile("/usr/local/bin/app", binary, ContainerWithFileOpts{
			Permissions: 0555,
		}).
		WithExposedPort(8080).
		WithEntrypoint([]string{"app"})
}
