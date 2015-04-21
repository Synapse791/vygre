package main

import (
    "os"
    "encoding/json"
    "log"
    "time"
    "github.com/samalba/dockerclient"
)

type Config struct {
    InstallDir      string                  `json:"install_dir"`
    DockerEndpoint  string                  `json:"docker_endpoint"`
    CheckInterval   time.Duration           `json:"check_interval"`
    Auth            dockerclient.AuthConfig `json:"auth"`
}

var config Config

func setConfig(filePath string) {

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        log.Print("Config file not found: " + filePath)
        configFile = "defaults"

        setDefaults()

        return
    }

    // TODO: replace check() with proper handling

    file, err := os.Open(filePath)
    check(err)

    decoder := json.NewDecoder(file)

    err2 := decoder.Decode(&config)
    check(err2)

    setDefaults()

    return

}

func setDefaults() {
    if config.InstallDir == "" {
        config.InstallDir = "/etc/vygre"
        log.Print("Set default for InstallDir")
    }

    if config.DockerEndpoint == "" {
        config.DockerEndpoint = "unix:///var/run/docker.sock"
        log.Print("Set default for DockerEndpoint")
    }

    if config.CheckInterval == 0 {
        config.CheckInterval = 3
        log.Print("Set default for CheckInterval")
    }
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}