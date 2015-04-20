package main

import (
    "os"
    "encoding/json"
)

var config Config

type Config struct {
    InstallDir      string
    DockerEndpoint  string
}

func setConfig(filePath string) {

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        PrintLine("Config file not found. Using defaults", 0)
        PrintBreak()

        setDefaults()

        return
    }

    file, err := os.Open(filePath)
    check(err)

    PrintLine("Found config file: " + filePath, 0)
    PrintBreak()

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
}

func PrintConfig() {

    PrintLine("Config (" + configFile + ")", 0)
    PrintLine("InstallDir: " + config.InstallDir, 1)
    PrintLine("DockerEndpoint: " + config.DockerEndpoint, 1)

}

func check(e error) {
    if e != nil {
        panic(e)
    }
}