package main

import "fmt"

func PrintLine(line string, indent int) {

    var space string

    switch indent {
        case 1:
            space = "  > "
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