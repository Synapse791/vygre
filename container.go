package main

import (
    "io/ioutil"
    "os"
    "encoding/json"
)

type ContainerInfo struct {
    Hostname    string  `json:'hostname'`
    Image       string  `json:'image'`
    Instances   int     `json:'instances'`
}

func ReadContainerFiles() ([]ContainerInfo, error) {

    configDir := config.InstallDir + "/conf.d/"

    // TODO: Check if directory exists

    fileList, _ := ioutil.ReadDir(configDir)

    var containerInfo []ContainerInfo

    for _, file := range fileList {
        var cont ContainerInfo
        data, _ := os.Open(configDir + file.Name())
        // TODO: error handling

        decoder := json.NewDecoder(data)
        decoder.Decode(&cont)

        containerInfo = append(containerInfo, cont)
    }

    return containerInfo, nil

}
