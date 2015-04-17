package main

import "fmt"

func PrintLine(line string, indent int) {

    var space string

    if indent == 1 {
        space = "  > "
    } else {
        space = ""
    }

    fmt.Println(space + line)

}