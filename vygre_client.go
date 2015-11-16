package main

import (
    "os"
    "errors"
    "fmt"
    "encoding/json"
    "io/ioutil"
    log "github.com/Sirupsen/logrus"
    "github.com/fsouza/go-dockerclient"
    "strings"
    "time"
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
    Name            string      `json:"container_name"`
    Instances       int         `json:"instances"`
    Image           string      `json:"image"`
    Ports           []string    `json:"ports"`
    Volumes         []string    `json:"volumes"`
    Environments    []string    `json:"environments"`
}

type VygreCreateOptions struct {
    Instances   int
    Options     *docker.CreateContainerOptions
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
        log.WithError(err).Fatal("failed to read vygre configuration")
    }

    var config VygreConfig

    if err := json.Unmarshal(rawConfig, &config); err != nil {
        log.WithError(err).Fatalf("invalid vygre configuration")
    }

    client.Config = config

}

func (client *VygreClient) CheckConfig() {
    // Checks if auth is set
    if client.Config.Auth != (docker.AuthConfiguration{}) {
        log.Info("docker authentication found")
        log.Info("validating docker authentication")
        if err := client.DockerClient.AuthCheck(&client.Config.Auth); err != nil {
            log.WithError(err).Fatal("docker authentication failed")
        }
        log.Info("authentication successful")
    }

    if client.Config.CheckInterval == 0 {
        log.WithError(errors.New("check_interval must be more than 0")).Fatal("invalid configuration")
    }

    log.Infof("check interval set to %d seconds", client.Config.CheckInterval)

}

func (client *VygreClient) ReadContainerConfig() {
    if _, err := os.Stat(containerConfigDir); os.IsNotExist(err) {
        log.Fatalf("vygre config directory not found at %s", containerConfigDir)
    }

    fileList, err := ioutil.ReadDir(containerConfigDir)
    if err != nil {
        log.WithError(err).Fatalf("failed to get file list from %s", containerConfigDir)
    }

    for _, fileName := range fileList {
        fullName := fmt.Sprintf("%s/%s", containerConfigDir, fileName.Name())

        log.Infof("reading %s", fullName)
        rawJson, err := ioutil.ReadFile(fullName)
        if err != nil {
            log.WithError(err).Fatalf("failed to decode %s", fullName)
        }

        var config VygreContainerConfig
        if err := json.Unmarshal(rawJson, &config); err != nil {
            log.WithError(errors.New(fmt.Sprintf("invalid JSON in %s", fileName))).Fatal("invalid configuration")
        }

        if err := CheckContainerConfig(config); err != nil {
            log.WithError(err).Fatal("invalid configuration")
        }

        client.ContainerConfigs = append(client.ContainerConfigs, &config)

        log.Infof("read %s successfully", fullName)

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
                }
            }
        }

        if len(containerConfig.Environments) > 0 {
            config.Env  =   containerConfig.Environments
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
        vygreOptions.Options    =   &options

        client.CreateOptions    =   append(client.CreateOptions, &vygreOptions)
    }
}

func (c *VygreClient) RunServer() {
    for _ = range time.Tick(c.Config.CheckInterval * time.Second) {
        // TODO Check each container has desired count
        // TODO Create any new containers that are required
        println("test")
    }
}

func (client *VygreClient) PrintVersion() {
    fmt.Printf("vygre %s\n", client.Version)
}

func CheckContainerConfig(c VygreContainerConfig) error {
    if c.Image == "" {
        return errors.New("image is required")
    }
    if c.Instances == 0 {
        return errors.New("instance is required and must be more than 0")
    }
    // TODO check if volume is present on host system
    // TODO check if port is free to bind to
    return nil
}
