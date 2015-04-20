package main

import (
    "io/ioutil"
    "os"
    "encoding/json"
)

type ContainerConfig struct {
    Hostname    string  `json:'hostname'`
    Image       string  `json:'image'`
    Instances   int     `json:'instances'`
}

func ReadContainerFiles() ([]ContainerConfig, error) {
    configDir := config.InstallDir + "/conf.d/"

    fileList, _ := ioutil.ReadDir(configDir)

    var containerInfo []ContainerConfig

    for _, file := range fileList {
        var cont ContainerConfig
        data, _ := os.Open(configDir + file.Name())

        decoder := json.NewDecoder(data)
        decoder.Decode(&cont)

        containerInfo = append(containerInfo, cont)
    }

    return containerInfo, nil

}
