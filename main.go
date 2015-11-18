package main

import (
    "os"
    "os/user"
    "flag"
    log "github.com/Sirupsen/logrus"
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
    vygre.CheckContainerConfig()

    if flags.ConfigCheck {
        vygre.Logger.Info("configuration ok")
        return
    }

    vygre.ProcessContainerConfig()
    vygre.UpdateImages()

    vygre.RunServer()

}
