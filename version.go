package main

import "fmt"

const currentVersion    =   "1.0"

func (client *VygreClient) PrintVersion() {
    fmt.Printf("vygre %s\n", client.Version)
}
