package main

import (
    "flag"
)

var configFile string

func main() {

    flag.StringVar(&configFile, "c", "/etc/vygre/config.json", "file path to the JSON config file")

    flag.Parse()

    setConfig(configFile)

    PrintConfig()

}