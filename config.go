package main

import (
    "os"
    "encoding/json"
)

var config Config

type Config struct {
    InstallDir      string  `json:'install_dir'`
    DockerEndpoint  string  `json:'docker_endpoint'`
    CheckInterval   int     `json:'check_interval'`
}

func setConfig(filePath string) {

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        configFile = "defaults"

        setDefaults()

        return
    }

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
    }

    if config.DockerEndpoint == "" {
        config.DockerEndpoint = "unix:///var/run/docker.sock"
    }

    if config.CheckInterval == 0 {
        config.CheckInterval = 3
    }
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}