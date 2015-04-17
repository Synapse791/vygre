package main

import (
    "os"
    "encoding/json"
)

type Config struct {
    InstallDir  string
}

func getConfig(filePath string) Config {

    file, err := os.Open(filePath)
    check(err)

    decoder := json.NewDecoder(file)

    conf := Config{}

    err2 := decoder.Decode(&conf)
    check(err2)

    return conf

}

func check(e error) {
    if e != nil {
        panic(e)
    }
}