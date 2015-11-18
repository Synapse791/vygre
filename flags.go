package main

import (
    "flag"
    "os/user"
    log "github.com/Sirupsen/logrus"
)

type FlagOpts struct {
    ConfigCheck     bool
    Help            bool
    VersionCheck    bool
}

var flags FlagOpts

func init() {
    currentUser, _ := user.Current()

    if currentUser.Uid != "0" { log.Fatal("vygre must be run as root") }

    flag.BoolVar(&flags.Help, "h", false, "print help information")
    flag.BoolVar(&flags.ConfigCheck, "t", false, "test configuration")
    flag.BoolVar(&flags.VersionCheck, "v", false, "print version information")

    flag.Parse()
}
