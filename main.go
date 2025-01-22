package main

import (
	"fmt"
	"os"

	"github.com/dsa-randomizer/cmd"
)

func help() {
	fmt.Println("Usage: dsa-randomizer <command>")
	fmt.Println("Commands:")
	for _, c := range cmd.Commands {
		fmt.Println("\t", c.Command() + ":", c.Help())
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 || args[0] == "help" {
		help()
	} else {
		found := false

		for _, c := range cmd.Commands {
			if c.Command() == args[0] {
				found = true
				c.Run(args[1:])
			}
		}

		if !found {
			help()
		}
	}
}