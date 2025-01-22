package cmd

import (
	"fmt"
)


type DbCommand struct {}

func (d *DbCommand) Command() string {
	return "db"
}

func (d *DbCommand) Help() string {
	return "Database commands"
}

func (d *DbCommand) Run(args []string) error {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("Usage: dsa-randomizer db <command>")
		fmt.Println("Commands:")
		fmt.Println("\tsetup: configures db for randomizer cli")
		fmt.Println("\treset: clears all tables, use -kp flag to keep problems in db")
		return nil
	} else {
		subcommand := args[0]

		switch subcommand {
			case "setup": 
				fmt.Println("setup invoked")
			case "reset":
				fmt.Println("reset invoked")
			default:
				fmt.Println("Unknown db subcommand")
				return nil 
		}
	}

	return nil
}