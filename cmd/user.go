package cmd

import (
	"fmt"
)

type UserCommand struct {}

func (u *UserCommand) Command() string {
	return "user"
}

func (u *UserCommand) Help() string {
	return "Commands relating to user"
}

func (u *UserCommand) Run(args []string) error {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("Usage: dsa-randomizer user <command>")
		fmt.Println("Commands:")
		fmt.Println("\ttimer: configure timer in hours of how long you have to complete code challenges")
		fmt.Println("\tstreak: Prints current challenge streak")
		fmt.Println("\thistory: Shows history of all problems solved")
		return nil
	} else {
		subcommand := args[0]

		switch subcommand {
			case "timer":
				fmt.Println("timer was invoked")
			case "streak":
				fmt.Println("streak was invoked")
			case "history":
				fmt.Println("history was invoked")
			default:
				fmt.Println("Unknown user subcommand")
				return nil
		}
	}

	return nil
}