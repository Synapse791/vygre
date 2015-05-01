package main

import "flag"

type FlagOpts struct {
    ConfigFilePath  string
    ConfigCheck     bool
}

var flags FlagOpts

func ParseFlags() {
    flag.StringVar(&flags.ConfigFilePath, "c", "defaults", "file path to the JSON config file")
    flag.BoolVar(&flags.ConfigCheck, "t", false, "test configuration")

    flag.Parse()
}