package main

import (
    "flag"
    "os/user"
    log "github.com/Sirupsen/logrus"
)

type FlagOpts struct {
    DebugMode    bool
    Example      bool
    Help         bool
    TestConfig   bool
    VersionCheck bool
}

var flags FlagOpts

func init() {
    currentUser, _ := user.Current()

    if currentUser.Uid != "0" { log.Fatal("vygre must be run as root") }

    flag.BoolVar(&flags.DebugMode,      "d", false, "enables debug logging")
    flag.BoolVar(&flags.Example,        "e", false, "prints configuration templates Exits with code 0")
    flag.BoolVar(&flags.Help,           "h", false, "prints help information        Exits with code 0")
    flag.BoolVar(&flags.TestConfig,     "t", false, "tests configuration            Exits with code 0 on success or code 1 on failure")
    flag.BoolVar(&flags.VersionCheck,   "v", false, "prints version information     Exits with code 0")

    flag.Parse()
}
