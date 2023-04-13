package main

type ContainerState string

const (
	Up   ContainerState = "Up"
	Down ContainerState = "Down"
)

type Container struct {
	State   ContainerState
	Name    string
	Compose *DockerCompose
}
