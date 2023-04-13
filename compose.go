package main

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Docker interface {
	ReloadDockerCompose() error
	GetDockerCompose(id string) *DockerCompose
	AllDockerCompose() []*DockerCompose
	StartDockerCompose(d *DockerCompose) error
	StopDockerCompose(d *DockerCompose) error
	ListenToDockerCompose(d *DockerCompose, onOutput func(line string) error) error
}

type docker struct {
	path          string
	dockerCompose map[string]*DockerCompose
}

func NewDocker(path string) Docker {
	return &docker{
		path:          path,
		dockerCompose: map[string]*DockerCompose{},
	}
}

type DockerCompose struct {
	ID   string
	Name string
	Path string
}

func (d *docker) ReloadDockerCompose() error {
	return filepath.WalkDir(".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		if !info.IsDir() && filepath.Ext(path) == ".yml" && strings.Contains(info.Name(), "docker-compose") {
			if !d.hasDockerCompose(path) {
				d.addDockerCompose(path, info.Name())
			}
		}

		return nil
	})
}

func (d *docker) hasDockerCompose(path string) bool {
	for _, d := range d.dockerCompose {
		if d.Path == path {
			return true
		}
	}
	return false
}

func (d *docker) addDockerCompose(path, name string) {
	u, _ := uuid.NewRandom()

	d.dockerCompose[u.String()] = &DockerCompose{
		ID:   u.String(),
		Path: path,
		Name: name,
	}
}

func (d *docker) GetDockerCompose(id string) *DockerCompose {
	return d.dockerCompose[id]
}

func (d *docker) AllDockerCompose() []*DockerCompose {
	res := []*DockerCompose{}
	for _, compose := range d.dockerCompose {
		res = append(res, compose)
	}

	return res
}

func (d *docker) StartDockerCompose(dc *DockerCompose) error {
	cmd := exec.Command(
		"docker",
		"compose",
		"-f",
		dc.Path,
		"up",
		"-d",
	)

	return cmd.Run()
}

func (d *docker) StopDockerCompose(dc *DockerCompose) error {
	cmd := exec.Command(
		"docker",
		"compose",
		"-f",
		dc.Path,
		"stop",
	)

	return cmd.Run()
}

func (d *docker) ListenToDockerCompose(dc *DockerCompose, onOutput func(line string) error) error {
	c := exec.Command(
		"docker",
		"compose",
		"-f",
		dc.Path,
		"logs",
		"-f",
	)
	out, err := c.StdoutPipe()
	if err != nil {
		return err
	}

	c.Start()

	buf := bufio.NewReader(out)
	if err != nil {
		log.Fatalf("Failed to listen in on docker-compose file: %s", err)
	}

	for {
		line, _, err := buf.ReadLine()
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if err = onOutput(string(line)); err != nil {
			return err
		}
	}

	return nil
}
