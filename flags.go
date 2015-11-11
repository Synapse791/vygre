package main

import "flag"

type FlagOpts struct {
    ConfigFilePath  string
    ConfigCheck     bool
}

var flags FlagOpts

func init() {
    flag.StringVar(&flags.ConfigFilePath, "c", "defaults", "file path to the JSON config file")
    flag.BoolVar(&flags.ConfigCheck, "t", false, "test configuration")
    flag.BoolVar(&flags.ConfigCheck, "version", false, "print version information")

    flag.Parse()
}