package main

import (
    "io/ioutil"
    "os"
    "encoding/json"
    "errors"
    "github.com/samalba/dockerclient"
)

type ContainerInfo struct {
    Hostname    string                                  `json:"hostname"`
    Image       string                                  `json:"image"`
    Instances   int                                     `json:"instances"`
    Env         []string                                `json:"env"`
    Volumes     []string                                `json:"volumes"`
    Ports       map[string][]dockerclient.PortBinding   `json:"ports"`
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

        err := CheckContainerInfo(file, cont)
        if err != nil {
            return nil, err
        }

        containerInfo = append(containerInfo, cont)
    }

    return containerInfo, nil

}

func CheckContainerInfo(f os.FileInfo, c ContainerInfo) error {
    if c.Image == "" {
        return errors.New(f.Name() + ": missing image definition")
    }

    //TODO: Check if image has version appended

    if c.Instances == 0 {
        return errors.New(f.Name() + ": missing instances definition")
    }
    return nil
}