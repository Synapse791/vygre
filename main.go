package main

import (
    "flag"
)

func main() {

    var configFile string
    flag.StringVar(&configFile, "c", "/etc/vygre/config.json", "file path to the JSON config file")

    flag.Parse()

    conf := getConfig(configFile)

    PrintConfig(configFile, conf)

}