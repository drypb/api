package main

import (
	"context"
	"log"
	"os"

	"dagger.io/dagger"
)

func main() {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	src := client.Host().Directory(".")

	build := client.Container().
		From("golang:1.22.6").
		WithDirectory("/app", src).
		WithWorkdir("/app").
		WithExec([]string{"go", "build", "-o", "bin/api", "./cmd/api"})

	test := build.WithExec([]string{"go", "test", "./..."})

	deploy := test.WithEntrypoint([]string{"./bin/api"})

	_, err = deploy.Export(context.Background(), "/tmp/api:latest.tar")
	if err != nil {
		log.Fatal(err)
	}
}
