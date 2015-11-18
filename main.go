package main

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
