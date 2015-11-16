package main

type FlagOpts struct {
    ConfigFilePath  string
    ConfigCheck     bool
    VersionCheck    bool
}

var flags FlagOpts

