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
)

type VygreConfig struct {
    LogLevel        string                      `json:"log_level"`
    CheckInterval   time.Duration               `json:"check_interval"`
    Auth            docker.AuthConfiguration    `json:"auth"`
    SMTP            VygreSMTPConfig             `json:"smtp"`
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

    if client.Config.CheckInterval == 0 {
        client.Logger.WithField("error", "check_interval must be more than 0").Fatal("invalid configuration")
    }

    // Checks if auth is set
    if client.Config.Auth != (docker.AuthConfiguration{}) {
        client.Logger.Info("authentication found")
        client.Logger.Debug("beginning authentication check")
        if err := client.DockerClient.AuthCheck(&client.Config.Auth); err != nil {
            client.Logger.WithField("error", err.Error()).Fatal("docker authentication failed")
        }
        client.Logger.Info("authentication check successful")
    }

    if client.Config.SMTP != (VygreSMTPConfig{}) {
        client.Logger.Info("smtp configuration found")
        client.Logger.Debug("beginning authentication check")
        if client.Config.SMTP.Port == 0 {
            client.Logger.Debug("using default port 587")
            client.Config.SMTP.Port = 587
        }
        if err := client.CheckSMTPConfig(); err != nil {
            client.Logger.Error(err)
            client.Logger.Fatalf("failed to authenticate against SMTP server %s", client.Config.SMTP.Host)
        }
        client.Logger.Info("smtp configuration valid")
    }

    client.Logger.Debugf("check interval set to %d seconds", client.Config.CheckInterval)

}

type VygreContainerConfig struct {
    Name            string              `json:"container_name"`
    Instances       int                 `json:"instances"`
    Image           string              `json:"image"`
    Ports           []string            `json:"ports"`
    Volumes         []string            `json:"volumes"`
    Environments    map[string]string   `json:"env"`
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