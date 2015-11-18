package main

import (
    "os"
    "fmt"
    "encoding/json"
    "io/ioutil"
    log "github.com/Sirupsen/logrus"
    "github.com/fsouza/go-dockerclient"
    "strings"
    "time"
    "regexp"
)

const dockerEndpoint        = "unix:///var/run/docker.sock"
const configFilePath        = "/etc/vygre/config.json"
const containerConfigDir    = "/etc/vygre/conf.d"

type VygreClient struct {
    DockerClient        *docker.Client
    Config              VygreConfig
    ContainerConfigs    []*VygreContainerConfig
    CreateOptions       []*VygreCreateOptions
    Version             string
}

type VygreConfig struct {
    CheckInterval   time.Duration               `json:"check_interval"`
    Auth            docker.AuthConfiguration    `json:"auth"`
}

type VygreContainerConfig struct {
    Name            string              `json:"container_name"`
    Instances       int                 `json:"instances"`
    Image           string              `json:"image"`
    Ports           []string            `json:"ports"`
    Volumes         []string            `json:"volumes"`
    Environments    map[string]string   `json:"env"`
}

type VygreCreateOptions struct {
    Instances   int
    State       VygreOptionsState
    Options     docker.CreateContainerOptions
}

type VygreOptionsState struct {
    Active      bool
    Attempts    int
}

func NewVygreClient() *VygreClient {
    dockerClient, err := docker.NewClient(dockerEndpoint)
    if err != nil {
        log.Fatal(err.Error())
    }
    return &VygreClient{Version: currentVersion, DockerClient: dockerClient}
}

func (client *VygreClient) ReadConfig() {

    log.Infof("reading configuration from %s", configFilePath)

    if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
        log.Fatalf("vygre config file not found at %s", configFilePath)
    }

    rawConfig, err := ioutil.ReadFile(configFilePath)
    if err != nil {
        log.Error(err.Error())
        log.Fatal("failed to read vygre configuration")
    }

    var config VygreConfig

    if err := json.Unmarshal(rawConfig, &config); err != nil {
        log.Error(err.Error())
        log.Fatal("invalid vygre configuration")
    }

    client.Config = config

}

func (client *VygreClient) CheckConfig() {
    // Checks if auth is set
    if client.Config.Auth != (docker.AuthConfiguration{}) {
        log.Info("docker authentication found")
        log.Info("validating docker authentication")
        if err := client.DockerClient.AuthCheck(&client.Config.Auth); err != nil {
            log.Error(err.Error())
            log.Fatal("docker authentication failed")
        }
        log.Info("authentication successful")
    }

    if client.Config.CheckInterval == 0 {
        log.Error("check_interval must be more than 0")
        log.Fatal("invalid configuration")
    }

    log.Infof("check interval set to %d seconds", client.Config.CheckInterval)

}

func (client *VygreClient) ReadContainerConfig() {
    if _, err := os.Stat(containerConfigDir); os.IsNotExist(err) {
        log.Fatalf("vygre config directory not found at %s", containerConfigDir)
    }

    fileList, err := ioutil.ReadDir(containerConfigDir)
    if err != nil {
        log.Error(err.Error())
        log.Fatalf("failed to get file list from %s", containerConfigDir)
    }

    for _, fileName := range fileList {
        fullName := fmt.Sprintf("%s/%s", containerConfigDir, fileName.Name())

        log.Infof("reading %s", fullName)
        rawJson, err := ioutil.ReadFile(fullName)
        if err != nil {
            log.Error(err.Error())
            log.Fatalf("failed to decode %s", fullName)
        }

        var config VygreContainerConfig
        if err := json.Unmarshal(rawJson, &config); err != nil {
            log.Error(fmt.Sprintf("invalid JSON in %s", fileName))
            log.Fatal("invalid configuration")
        }

        client.ContainerConfigs = append(client.ContainerConfigs, &config)

        log.Infof("read %s successfully", fullName)

    }

}

func (client *VygreClient) CheckContainerConfig() {
    for _, config := range client.ContainerConfigs {
        if config.Instances < 1 {
            log.Fatalf("instances must be at least 1: '%d' given", config.Instances)
        }
        if match, _ := regexp.MatchString("^[a-zA-Z0-9_]{4,}$", config.Name); config.Name != "" && !match {
            log.Fatal("container name must contain only alphanumeric and underscores and be at least 4 characters long")
        }
        if match, _ := regexp.MatchString("^(?:(?:[a-zA-Z0-9-.:]+)+/)?(?:[a-zA-Z0-9-]+/)?[a-zA-Z0-9-]+(?::[a-zA-Z0-9-.]+)?$", config.Image); !match {
            log.Fatal("image must be a standard docker image name with option registry location and/or tag")
        }
        for _, port := range config.Ports {
            if match, _ := regexp.MatchString("^(?:[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:)?(?:[0-9]{1,5}:)?[0-9]{1,5}$", port); !match {
                log.Error("ports must follow the format of the docker run -p flag (https://docs.docker.com/engine/reference/run/#expose-incoming-ports)")
                log.Fatalf("invalid port '%s'", port)
            }
        }
        for _, volume := range config.Volumes {
            parts := strings.Split(volume, ":")
            if _, err := os.Stat(parts[0]); os.IsNotExist(err) {
                log.Error(err.Error())
                log.Fatal("volume mount not found")
            }
            if match, _ := regexp.MatchString("^[/a-zA-z0-9-_\\.]+:[/a-zA-z0-9-_\\.]+", volume); !match {
                log.Fatal("image must be a standard docker image name with option registry location and/or tag")
            }
        }
    }
}

func (client *VygreClient) ProcessContainerConfig() {
    log.Infof("processing %d configuration(s)", len(client.ContainerConfigs))

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
            for _, port := range containerConfig.Ports {
                if strings.Contains(port, ":") {
                    var binding docker.PortBinding

                    parts   := strings.Split(port, ":")

                    if strings.Contains(parts[0], ".") {
                        binding.HostIP      =   parts[0]
                        binding.HostPort    =   parts[1]
                        hostConfig.PortBindings[docker.Port(parts[2])]  =   append(hostConfig.PortBindings[docker.Port(parts[2])], binding)
                    } else {
                        binding.HostIP      =   "0.0.0.0"
                        binding.HostPort    =   parts[0]
                        bindingMap  := make(map[docker.Port][]docker.PortBinding)

                        bindingMap[docker.Port(parts[1] + "/tcp")] = []docker.PortBinding{binding}
                        hostConfig.PortBindings =   bindingMap
                    }
                } else {
                    portMap := make(map[docker.Port]struct{})

                    var empty struct{}

                    portMap[docker.Port(fmt.Sprintf("%s/tcp", port))]   =   empty

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
                var mount docker.Mount
                parts               := strings.Split(volume, ":")
                mount.Source        = parts[0]
                mount.Destination   = parts[1]
                if len(parts) > 2 {
                    if parts[2] == "ro" {
                        mount.Mode  =   "ro"
                        mount.RW    =   false
                    } else {
                        mount.Mode  =   "rw"
                        mount.RW    =   true
                    }
                }
                config.Mounts   =   append(config.Mounts, mount)
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

        log.Infof("pulling %s image", config.Image)

        var auth docker.AuthConfiguration

        if pullOptions.Registry != "" && strings.Contains(client.Config.Auth.ServerAddress, pullOptions.Registry) {
            log.Info("using auth")
            auth = client.Config.Auth
        }

        if err := client.DockerClient.PullImage(pullOptions, auth); err != nil {
            log.Error(err.Error())
            log.Fatal("failed to pull docker image")
        }

        log.Info("pulled successfully")

    }
}

func (client *VygreClient) RunServer() {
    for _ = range time.Tick(client.Config.CheckInterval * time.Second) {
        log.Info("checking running containers")

        for _, options := range client.CreateOptions {
            if ! options.State.Active {
                log.Warnf("INACTIVE: %s", options.Options.Config.Image)
                continue
            }

            count := client.GetContainerCount(options.Options.Config.Image)

            log.Infof("%d of %d %s containers running", count, options.Instances, options.Options.Config.Image)

            if count < options.Instances {
                log.Infof("creating new %s container", options.Options.Config.Image)
                options.Options.Name = ""
                new, err := client.DockerClient.CreateContainer(options.Options)
                if err != nil {
                    log.Error(err.Error())
                    log.Fatal("failed to create container")
                }
                log.Infof("created %s", new.ID)

                log.Infof("starting %s", new.ID)
                if err := client.DockerClient.StartContainer(new.ID, new.HostConfig); err != nil {
                    log.Error(err.Error())
                    log.Fatal("failed to start container")
                }

                time.Sleep(2 * time.Second)

                newCount := client.GetContainerCount(options.Options.Config.Image)

                if newCount != count + 1 {
                    log.Warnf("failed to start %s", new.ID)
                    options.State.Attempts++
                    if options.State.Attempts > 3 {
                        options.State.Active    =   false
                        log.Errorf("setting %s to INACTIVE", options.Options.Config.Image)
                    }
                }
            }
        }
    }
}

func (client *VygreClient) PrintVersion() {
    fmt.Printf("vygre %s\n", client.Version)
}

func (client *VygreClient) GetContainerCount(image string) int {
    containerList, err := client.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
    if err != nil {
        log.Error(err.Error())
        log.Fatal("failed to list running containers")
    }

    count := 0

    for _, container := range containerList {
        if container.Image == image {
            count++
        }
    }

    return count
}
