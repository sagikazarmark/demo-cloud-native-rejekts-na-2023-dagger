package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

const (
	goVersion           = "1.21.3"
	golangciLintVersion = "1.54.2"
	artifactBaseImage   = "alpine:3.18.4"
)

type Ci struct{}

// Run the entire CI pipeline
func (m *Ci) CI(ctx context.Context, githubRelease Optional[bool], version Optional[string], githubActor Optional[string], githubToken Optional[*Secret]) error {
	var group errgroup.Group

	appVersion := version.GetOr("latest")

	// Build
	var app *Container
	{
		dag := dag.Pipeline("Build")

		group.Go(func() error {
			var err error

			app, err = build(dag, appVersion).Sync(ctx)

			return err
		})
	}

	// Test
	{
		dag := dag.Pipeline("Test")

		group.Go(func() error {
			_, err := test(dag).Sync(ctx)

			return err
		})
	}

	// Lint
	{
		dag := dag.Pipeline("Lint")

		group.Go(func() error {
			_, err := lint(dag).Sync(ctx)

			return err
		})
	}

	err := group.Wait()
	if err != nil {
		return err
	}

	// TODO: remove this from the initial version and add back later
	_, err = dag.Trivy().ScanContainer(ctx, app, TrivyScanContainerOpts{
		Severity: "HIGH,CRITICAL",
		ExitCode: 1,
	})
	if err != nil {
		return err
	}

	// If this is a release, publish and deploy the container image
	if githubRelease.GetOr(false) {
		username, ok := githubActor.Get()
		if !ok {
			return fmt.Errorf("github actor not set")
		}

		token, ok := githubToken.Get()
		if !ok {
			return fmt.Errorf("github token not set")
		}

		const imageRepo = "ghcr.io/sagikazarmark/demo-cloud-native-rejekts-na-2023-dagger"

		ref, err := app.WithRegistryAuth("ghcr.io", username, token).Publish(ctx, fmt.Sprintf("%s:%s", imageRepo, appVersion))
		if err != nil {
			return err
		}

		// Deploy ref
		// TODO(mark): deploy to Fly.io
		_ = ref
	}

	return nil
}

// Build the container image
func (m *Ci) Build() *Container {
	return build(dag, "latest")
}

func build(dag *Client, version string) *Container {
	host := dag.Host()

	binary := dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedDirectory("/src", host.Directory(root(), HostDirectoryOpts{
			Exclude: []string{".direnv", ".devenv"},
		})).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOOS", "linux").
		WithEnvVariable("GOARCH", "amd64").
		WithEnvVariable("VERSION", version).
		WithExec([]string{"go", "build", "-ldflags", "-X main.version=${VERSION}", "-o", "/usr/local/bin/app", "."}).
		File("/usr/local/bin/app")

	return dag.Container().From(artifactBaseImage).
		WithExec([]string{"apk", "add", "--update", "--no-cache", "ca-certificates", "tzdata"}).
		WithFile("/usr/local/bin/app", binary, ContainerWithFileOpts{
			Permissions: 0555,
		}).
		WithExposedPort(8080).
		WithEntrypoint([]string{"app"})
}

// Test jobs
func (m *Ci) Test() *Container {
	return test(dag)
}

func test(dag *Client) *Container {
	host := dag.Host()

	return dag.Container().
		From(fmt.Sprintf("golang:%s", goVersion)).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedDirectory("/src", host.Directory(root(), HostDirectoryOpts{
			Exclude: []string{".direnv", ".devenv"},
		})).
		WithWorkdir("/src").
		WithExec([]string{"go", "test", "-v", "./..."})
}

// Linter jobs
func (m *Ci) Lint() *Container {
	return lint(dag)
}

func lint(dag *Client) *Container {
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
