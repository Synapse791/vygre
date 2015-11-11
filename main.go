package main

import (
    "log"
    "time"
)

func main() {

    if flags.VersionCheck {
        PrintVersion()
        return
    }

    setConfig(flags.ConfigFilePath)

    if flags.ConfigCheck {
        PrintConfig()
        return
    }

    containerConfigs, err := ReadContainerFiles()
    if err != nil {
        log.Print("config_check: FAILED")
        log.Fatal(err)
    }

    log.Print("config_check: PASSED")

    CheckContainers(containerConfigs)

    for _ = range time.Tick(config.CheckInterval * time.Second) {
        CheckContainers(containerConfigs)
    }

}

func CheckContainers(containerConfigs []ContainerInfo) {
    log.Print("------------------------------------------------")

    for _, c := range containerConfigs {
        count, _ := GetContainerCount(c.Image)
        log.Printf("%s > %d of %d running", c.Image, count, c.Instances)
        if count < c.Instances {
            log.Printf("increasing %s count by %d", c.Image, c.Instances - count)
            for loopCount := count; loopCount < c.Instances; loopCount ++{
                CreateContainer(c)
            }
        }
    }
}