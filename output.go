package main

import "fmt"

func PrintLine(line string, indent int) {

    var space string

    switch indent {
        case 1:
            space = "|-> "
        case 2:
            space = "  |-> "
        default:
            space = ""
    }

    fmt.Println(space + line)

}

func PrintBreak() {
    fmt.Println("")
}

func PrintConfig() {
    chk := fmt.Sprintf("CheckInterval: %d seconds", config.CheckInterval)

    PrintLine("--------------------------------------",0)
    PrintLine("Config (" + configFile + ")", 0)
    PrintLine("InstallDir: " + config.InstallDir, 1)
    PrintLine("DockerEndpoint: " + config.DockerEndpoint, 1)
    PrintLine(chk, 1)
    PrintLine("--------------------------------------",0)
}