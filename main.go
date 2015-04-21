package main

import (
    "flag"
    "log"
    "time"
)

var configFile string

func main() {

    var isTest bool

    flag.StringVar(&configFile, "c", "/etc/vygre/config.json", "file path to the JSON config file")
    flag.BoolVar(&isTest, "t", false, "test configuration")

    flag.Parse()

    setConfig(configFile)

    if isTest {
        PrintConfig()
        return
    }

    containerConfigs, err := ReadContainerFiles()
    if err != nil {
        log.Print("config_check: FAILED")
        log.Fatal(err)
    }

    log.Print("config_check: PASSED")

    for _ = range time.Tick(config.CheckInterval * time.Second) {
        CheckContainers(containerConfigs)
    }

}

func CheckContainers(containerConfigs []ContainerInfo) {
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