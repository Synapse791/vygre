package main

import (
    "github.com/samalba/dockerclient"
    "log"
)

func GetContainerCount(image string) (int, error) {
    docker, _ := dockerclient.NewDockerClient(config.DockerEndpoint, nil)

    containers, _ := docker.ListContainers(false, false, "")

    count := 0

    for _, c := range containers {
        if c.Image == image {
            count++
        }
    }

    return count, nil
}

func CreateContainer(c ContainerInfo) {
    var createInfo dockerclient.ContainerConfig
    var hostConfig dockerclient.HostConfig

    createInfo.Image = c.Image

    //TODO: Check is image exists. Pull if not

    if c.Hostname != "" {
        createInfo.Hostname = c.Hostname
    }

    if c.Env != nil {
        createInfo.Env = c.Env
    }

    if c.Volumes != nil {
        hostConfig.Binds = c.Volumes
    }

    if c.Ports != nil {
        hostConfig.PortBindings = c.Ports
    } else {
        hostConfig.PublishAllPorts = true
    }

    log.Print("creating new " + createInfo.Image)

    docker, _ := dockerclient.NewDockerClient(config.DockerEndpoint, nil)

    id, err := docker.CreateContainer(&createInfo, "")
    if err != nil {
        log.Fatal(err)
    }

    log.Print("starting ", id)

    err2 := docker.StartContainer(id, &hostConfig)
    if err2 != nil {
        log.Fatal(err2)
    }

}
