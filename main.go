package main

import (
	"log"
	"os"
)

func main() {
	currentWorkDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %s", err)
	}
	docker := NewDocker(currentWorkDir)
	if err := docker.ReloadDockerCompose(); err != nil {
		log.Fatalf("Failed to load docker-compose files: %s", err)
	}

	someDC := docker.AllDockerCompose()[0]
	if err := docker.StartDockerCompose(someDC); err != nil {
		log.Fatalf("Failed to start docker-compose file: %s", err)
	}

	if err := docker.ListenToDockerCompose(someDC, func(line string) error {
		log.Println(line)
		return nil
	}); err != nil {
		log.Fatalf("Something went wrong while listening in on docker-compose file: %s", err)
	}
}
