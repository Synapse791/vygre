package main

import (
    "fmt"
    "github.com/fsouza/go-dockerclient"
    "strings"
    "time"
)

type VygreCreateOptions struct {
    Instances   int
    State       VygreOptionsState
    Options     docker.CreateContainerOptions
}

type VygreOptionsState struct {
    Active      bool
    Attempts    int
}

func (client *VygreClient) ProcessContainerConfig() {
    client.Logger.Debugf("processing %d configuration(s)", len(client.ContainerConfigs))

    for _, containerConfig := range client.ContainerConfigs {
        var vygreOptions    VygreCreateOptions
        var options         docker.CreateContainerOptions
        var config          docker.Config
        var hostConfig      docker.HostConfig

        vygreOptions.Instances  =   containerConfig.Instances
        config.Image    =   containerConfig.Image

        if containerConfig.Name != "" {
            options.Name    =   containerConfig.Name
        }

        if len(containerConfig.Ports) > 0 {
            hostConfig.NetworkMode  =   "default"
            for _, port := range containerConfig.Ports {
                if strings.Contains(port, ":") {
                    var binding docker.PortBinding

                    parts   := strings.Split(port, ":")

                    if strings.Contains(parts[0], ".") {
                        binding.HostIP      =   parts[0]
                        binding.HostPort    =   parts[1]
                        hostConfig.PortBindings[docker.Port(parts[2])]  =   append(hostConfig.PortBindings[docker.Port(parts[2])], binding)
                        portMap := map[docker.Port]struct{}{
                            docker.Port(fmt.Sprintf("%s/tcp", parts[2])): {},
                        }
                        config.ExposedPorts = portMap
                    } else {
                        binding.HostIP      =   ""
                        binding.HostPort    =   parts[0]
                        bindingMap  := make(map[docker.Port][]docker.PortBinding)

                        bindingMap[docker.Port(parts[1] + "/tcp")] = []docker.PortBinding{binding}
                        hostConfig.PortBindings =   bindingMap
                        portMap := map[docker.Port]struct{}{
                            docker.Port(fmt.Sprintf("%s/tcp", parts[1])): {},
                        }
                        config.ExposedPorts = portMap
                    }
                } else {
                    hostConfig.PublishAllPorts  =   true

                    portMap := map[docker.Port]struct{}{
                        docker.Port(fmt.Sprintf("%s/tcp", port)): {},
                    }
                    config.ExposedPorts = portMap
                }

            }
        }

        if len(containerConfig.Environments) > 0 {
            for key, value := range containerConfig.Environments {
                combined    :=  fmt.Sprintf("%s=%s", key, value)
                config.Env  =   append(config.Env, combined)
            }
        }

        if len(containerConfig.Volumes) > 0 {
            for _, volume := range containerConfig.Volumes {
                hostConfig.Binds    =   append(hostConfig.Binds, volume)
            }
        }

        options.Config          =   &config
        options.HostConfig      =   &hostConfig

        vygreOptions.State.Active   =   true
        vygreOptions.State.Attempts =   0
        vygreOptions.Options        =   options

        client.CreateOptions    =   append(client.CreateOptions, &vygreOptions)
    }
}

func (client *VygreClient) UpdateImages() {
    for _, config := range client.ContainerConfigs {
        var pullOptions docker.PullImageOptions

        parts := strings.Split(config.Image, "/")

        if strings.Contains(parts[0], ".") {
            pullOptions.Registry    =   parts[0]
        }

        if strings.Contains(parts[len(parts) - 1], ":") {
            tagParts        :=  strings.Split(parts[len(parts) - 1], ":")
            pullOptions.Tag =   tagParts[1]
        } else {
            pullOptions.Tag =   "latest"
        }

        suffixTrim  :=  strings.TrimSuffix(config.Image, fmt.Sprintf(":%s", pullOptions.Tag))

        pullOptions.Repository  =   suffixTrim

        client.Logger.Infof("pulling %s image", config.Image)

        var auth docker.AuthConfiguration

        if pullOptions.Registry != "" && strings.Contains(client.Config.Auth.ServerAddress, pullOptions.Registry) {
            client.Logger.Debug("using auth")
            auth = client.Config.Auth
        }

        if err := client.DockerClient.PullImage(pullOptions, auth); err != nil {
            client.Logger.WithField("error", err.Error()).Fatal("failed to pull docker image")
        }

        client.Logger.Info("pull complete")

    }
}

func (client *VygreClient) RunServer() {
    for _ = range time.Tick(client.Config.CheckInterval * time.Second) {
        client.Logger.Debug("checking required containers are running")

        for _, options := range client.CreateOptions {
            if ! options.State.Active {
                client.Logger.Debugf("INACTIVE: %s", options.Options.Config.Image)
                continue
            }

            count := client.GetContainerCount(options.Options.Config.Image)

            client.Logger.Debugf("%d of %d %s containers running", count, options.Instances, options.Options.Config.Image)

            if count < options.Instances {
                client.Logger.Infof("creating new %s container", options.Options.Config.Image)
                options.Options.Name = ""
                new, err := client.DockerClient.CreateContainer(options.Options)
                if err != nil {
                    client.Logger.WithField("error", err.Error()).Fatal("failed to create container")
                }
                client.Logger.Infof("created %s", new.ID[0:8])

                client.Logger.Infof("starting %s", new.ID[0:8])
                if err := client.DockerClient.StartContainer(new.ID, new.HostConfig); err != nil {
                    client.Logger.WithField("error", err.Error()).Fatal("failed to start container")
                }

                time.Sleep(2 * time.Second)

                newCount := client.GetContainerCount(options.Options.Config.Image)

                if newCount != count + 1 {
                    client.Logger.Warnf("failed to start %s", new.ID)
                    options.State.Attempts++
                    if options.State.Attempts > 2 {
                        options.State.Active    =   false
                        client.Logger.Errorf("setting %s to INACTIVE after 3 failed attempts", options.Options.Config.Image)
                        client.Logger.Warn("sending INACTIVE alert email")
                        client.SendInactiveNotification(options.Options.Config.Image)
                    }
                } else {
                    client.Logger.Infof("started %s", new.ID[0:8])
                }
            }
        }
    }
}

func (client *VygreClient) GetContainerCount(image string) int {
    containerList, err := client.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
    if err != nil {
        client.Logger.WithField("error", err.Error()).Fatal("failed to list running containers")
    }

    count := 0

    for _, container := range containerList {
        if container.Image == image {
            count++
        }
    }

    return count
}