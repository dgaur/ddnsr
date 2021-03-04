package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Parse()
	for _, host := range flag.Args() {
		fmt.Println(host)
		resolve(host)
	}

	return
}
