package main
import (
    "fmt"
    "flag"
)

const VYGRE_HELP_TEXT = `Usage: vygre [OPTIONS]

A service for starting new docker containers when the old ones die

Options:
`

const VYGRE_CONFIGURATION_TEMPLATE = `{
  "log_level"       : "",
  "check_interval"  : 0,
  "auth"            : {
    "serveraddress"   : "",
    "username"        : "",
    "password"        : "",
    "email"           : ""
  },
  "smtp" : {
    "host"     : "",
    "user"     : "",
    "password" : "",
    "to"       : ""
  }
}`

const VYGRE_CONTAINER_CONFIGURATION_TEMPLATE = `{
  "image"     : "",
  "instances" : 0,
  "env"       : {
    "" : ""
  },
  "ports": [ "" ],
  "volumes" : [ "" ]
}`

func (client *VygreClient) PrintHelp() {
    fmt.Print(VYGRE_HELP_TEXT)
    flag.PrintDefaults()
}

func (client *VygreClient) PrintConfigurationTemplates() {
    fmt.Println("--------------------------------------")
    fmt.Println(" /etc/vygre/config.json")
    fmt.Println("--------------------------------------")
    fmt.Println(VYGRE_CONFIGURATION_TEMPLATE)
    fmt.Println("--------------------------------------")
    fmt.Println(" /etc/vygre/conf.d/*")
    fmt.Println("--------------------------------------")
    fmt.Println(VYGRE_CONTAINER_CONFIGURATION_TEMPLATE)
    fmt.Println("--------------------------------------")
}
