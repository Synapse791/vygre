package main

import (
    "os"
    "fmt"
    "github.com/Sirupsen/logrus"
    "github.com/fsouza/go-dockerclient"
    "flag"
)

type VygreClient struct {
    Logger              *logrus.Logger
    DockerClient        *docker.Client
    Config              VygreConfig
    ContainerConfigs    []*VygreContainerConfig
    CreateOptions       []*VygreCreateOptions
    Version             string
}

func NewVygreClient() *VygreClient {

    logger          := logrus.New()
    logger.Out      =   os.Stdout
    logger.Level    =   logrus.InfoLevel

    dockerClient, err := docker.NewClient(DOCKER_ENDPOINT)
    if err != nil {
        logger.WithField("error", err.Error()).Fatal("failed to create docker client")
    }

    return &VygreClient{Version: currentVersion, DockerClient: dockerClient, Logger: logger}
}

func (client *VygreClient) PrintHelp() {
    fmt.Print(VYGRE_HELP_TEXT)
    flag.PrintDefaults()
}

func (client *VygreClient) PrintVersion() {
    fmt.Printf("vygre %s\n", client.Version)
}

