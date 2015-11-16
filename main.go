package main

import (
    "flag"
    "os/user"
    "os"
    log "github.com/Sirupsen/logrus"
    "encoding/json"
)

var vygre *VygreClient

func init() {
    currentUser, _ := user.Current()

    if currentUser.Uid != "0" { log.Fatal("vygre must be run as root") }

    flag.StringVar(&flags.ConfigFilePath, "c", "defaults", "file path to the JSON config file")
    flag.BoolVar(&flags.ConfigCheck, "t", false, "test configuration")
    flag.BoolVar(&flags.VersionCheck, "v", false, "print version information")

    flag.Parse()
}

func main() {

    vygre = NewVygreClient()

    if flags.VersionCheck {
        vygre.PrintVersion()
        os.Exit(0)
    }

    vygre.ReadConfig()
    vygre.CheckConfig()
    vygre.ReadContainerConfig()
    vygre.ProcessContainerConfig()

//    for _ = range time.Tick(config.CheckInterval * time.Second) {
//        CheckContainers(containerConfigs)
//    }

}
