package main

const DOCKER_ENDPOINT       =   "unix:///var/run/docker.sock"
const CONFIG_FILE_PATH      =   "/etc/vygre/config.json"
const CONTAINER_CONFIG_DIR  =   "/etc/vygre/conf.d"

var vygre *VygreClient

func main() {

    vygre = NewVygreClient()

    if flags.VersionCheck {
        vygre.PrintVersion()
        return
    }
    if flags.Help {
        vygre.PrintHelp()
        return
    }
    if flags.Example {
        vygre.PrintConfigurationTemplates()
        return
    }

    vygre.Logger.Warn("initializing server")

    vygre.ReadConfig()
    vygre.CheckConfig()
    vygre.ReadContainerConfig()
    vygre.CheckContainerConfig()

    if flags.TestConfig {
        vygre.Logger.Info("configuration ok")
        return
    }

    vygre.ProcessContainerConfig()
    vygre.UpdateImages()

    vygre.Logger.Warn("initialization complete")
    vygre.Logger.Warn("running server")

    vygre.RunServer()

}
