package main

import (
    "os"
    "fmt"
    "encoding/json"
    "io/ioutil"
    "github.com/Sirupsen/logrus"
    "github.com/fsouza/go-dockerclient"
    "strings"
    "time"
    "regexp"
    "flag"
)

const DOCKER_ENDPOINT       =   "unix:///var/run/docker.sock"
const CONFIG_FILE_PATH      =   "/etc/vygre/config.json"
const CONTAINER_CONFIG_DIR  =   "/etc/vygre/conf.d"

type VygreClient struct {
    Logger              *logrus.Logger
    DockerClient        *docker.Client
    Config              VygreConfig
    ContainerConfigs    []*VygreContainerConfig
    CreateOptions       []*VygreCreateOptions
    Version             string
}

type VygreConfig struct {
    LogLevel        string                      `json:"log_level"`
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

var vygre *VygreClient

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

func (client *VygreClient) ReadConfig() {

    client.Logger.Debugf("reading configuration from %s", CONFIG_FILE_PATH)

    if _, err := os.Stat(CONFIG_FILE_PATH); os.IsNotExist(err) {
        client.Logger.Fatalf("vygre config file not found at %s", CONFIG_FILE_PATH)
    }

    rawConfig, err := ioutil.ReadFile(CONFIG_FILE_PATH)
    if err != nil {
        client.Logger.WithField("error", err.Error()).Fatal("failed to read vygre configuration")
    }

    var config VygreConfig

    if err := json.Unmarshal(rawConfig, &config); err != nil {
        client.Logger.WithField("error", err.Error()).Fatalf("invalid vygre configuration")
    }

    client.Config = config

}

func (client *VygreClient) CheckConfig() {
    if flags.DebugMode {
        client.Logger.Level =   logrus.DebugLevel
    } else if client.Config.LogLevel != "" {
        switch client.Config.LogLevel {
        case "debug":
            client.Logger.Level =   logrus.DebugLevel
            break
        case "info":
            client.Logger.Level =   logrus.InfoLevel
            break
        case "warning":
            client.Logger.Level =   logrus.WarnLevel
            break
        case "error":
            client.Logger.Level =   logrus.ErrorLevel
            break
        }
    }

    client.Logger.Debug("debug logging enabled")

    // Checks if auth is set
    if client.Config.Auth != (docker.AuthConfiguration{}) {
        client.Logger.Info("authentication found")
        client.Logger.Debug("beginning authentication check")
        if err := client.DockerClient.AuthCheck(&client.Config.Auth); err != nil {
            client.Logger.WithField("error", err.Error()).Fatal("docker authentication failed")
        }
        client.Logger.Info("authentication check successful")
    }

    if client.Config.CheckInterval == 0 {
        client.Logger.WithField("error", "check_interval must be more than 0").Fatal("invalid configuration")
    }

    client.Logger.Debugf("check interval set to %d seconds", client.Config.CheckInterval)

}

func (client *VygreClient) ReadContainerConfig() {
    client.Logger.Info("reading container configuration")

    if _, err := os.Stat(CONTAINER_CONFIG_DIR); os.IsNotExist(err) {
        client.Logger.Fatalf("vygre config directory not found at %s", CONTAINER_CONFIG_DIR)
    }

    fileList, err := ioutil.ReadDir(CONTAINER_CONFIG_DIR)
    if err != nil {
        client.Logger.WithField("error", err.Error()).Fatalf("failed to get file list from %s", CONTAINER_CONFIG_DIR)
    }

    for _, fileName := range fileList {
        fullName := fmt.Sprintf("%s/%s", CONTAINER_CONFIG_DIR, fileName.Name())

        client.Logger.Debugf("reading %s", fullName)
        rawJson, err := ioutil.ReadFile(fullName)
        if err != nil {
            client.Logger.WithField("error", err.Error()).Fatalf("failed to decode %s", fullName)
        }

        var config VygreContainerConfig
        if err := json.Unmarshal(rawJson, &config); err != nil {
            client.Logger.WithField("error", fmt.Sprintf("invalid JSON in %s", fileName)).Fatal("invalid configuration")
        }

        client.ContainerConfigs = append(client.ContainerConfigs, &config)

        client.Logger.Debugf("read %s successfully", fullName)

    }

}

func (client *VygreClient) CheckContainerConfig() {
    for _, config := range client.ContainerConfigs {
        if config.Instances < 1 {
            client.Logger.Fatalf("instances must be at least 1: '%d' given", config.Instances)
        }
        if match, _ := regexp.MatchString("^[a-zA-Z0-9_]{4,}$", config.Name); config.Name != "" && !match {
            client.Logger.Fatal("container name must contain only alphanumeric and underscores and be at least 4 characters long")
        }
        if match, _ := regexp.MatchString("^(?:(?:[a-zA-Z0-9-.:]+)+/)?(?:[a-zA-Z0-9-]+/)?[a-zA-Z0-9-]+(?::[a-zA-Z0-9-.]+)?$", config.Image); !match {
            client.Logger.Fatal("image must be a standard docker image name with option registry location and/or tag")
        }
        for _, port := range config.Ports {
            if match, _ := regexp.MatchString("^(?:[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}:)?(?:[0-9]{1,5}:)?[0-9]{1,5}$", port); !match {
                client.Logger.WithField("error", "ports must follow the format of the docker run -p flag (https://docs.docker.com/engine/reference/run/#expose-incoming-ports)").Fatalf("invalid port '%s'", port)
            }
        }
        for _, volume := range config.Volumes {
            parts := strings.Split(volume, ":")
            if _, err := os.Stat(parts[0]); os.IsNotExist(err) {
                client.Logger.WithField("error", err.Error()).Fatalf("volume mount not found")
            }
            if match, _ := regexp.MatchString("^[/a-zA-z0-9-_\\.]+:[/a-zA-z0-9-_\\.]+", volume); !match {
                client.Logger.Fatal("image must be a standard docker image name with option registry location and/or tag")
            }
        }
    }
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
                    if options.State.Attempts > 3 {
                        options.State.Active    =   false
                        client.Logger.Errorf("setting %s to INACTIVE", options.Options.Config.Image)
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

func (client *VygreClient) PrintHelp() {
    fmt.Print(VYGRE_HELP_TEXT)
    flag.PrintDefaults()
}

func (client *VygreClient) PrintVersion() {
    fmt.Printf("vygre %s\n", client.Version)
}

