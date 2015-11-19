package main

import (
    "os"
    "github.com/Sirupsen/logrus"
    "github.com/fsouza/go-dockerclient"
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
