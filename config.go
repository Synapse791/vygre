package main

import (
    "os"
    "encoding/json"
)

type Config struct {
    InstallDir      string
    DockerEndpoint  string
}

func getConfig(filePath string) Config {

    conf := Config{}

    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        PrintLine("Config file not found. Using defaults", 0)
        PrintBreak()

        setDefaults(&conf)

        return conf
    }

    file, err := os.Open(filePath)
    check(err)

    PrintLine("Found config file: " + filePath, 0)
    PrintBreak()

    decoder := json.NewDecoder(file)

    err2 := decoder.Decode(&conf)
    check(err2)

    setDefaults(&conf)

    return conf

}

func setDefaults(conf *Config) {
    if conf.InstallDir == "" {
        conf.InstallDir = "/etc/vygre"
    }

    if conf.DockerEndpoint == "" {
        conf.DockerEndpoint = "unix:///var/run/docker.sock"
    }
}

func PrintConfig(configFile string, conf Config) {

    PrintLine("Config (" + configFile + ")", 0)
    PrintLine("InstallDir: " + conf.InstallDir, 1)
    PrintLine("DockerEndpoint: " + conf.DockerEndpoint, 1)

}

func check(e error) {
    if e != nil {
        panic(e)
    }
}