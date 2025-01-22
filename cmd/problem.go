package cmd

import (
	"fmt"
)


type ProblemCommand struct {}

func (p *ProblemCommand) Command() string {
	return "problem"
}

func (p *ProblemCommand) Help() string {
	return "Commands relating to code problems and challenges"
}

func (p *ProblemCommand) Run(args []string) error {
	if len(args) == 0 || args[0] == "help" {
		fmt.Println("Usage: dsa-randomizer problem <command>")
		fmt.Println("Commands:")
		fmt.Println("\tadd: adds problem to the db, flags: -n name -l link")
		fmt.Println("\tdelete: deletes problem from the db, flags: -id id")
		fmt.Println("\tstart: provides a random problem and starts a new timer, if a problem is currently open then it closes the timer and ends your streak")
		fmt.Println("\tdone: marks current open problem to done, boosts streak if within timer set on the user")
		fmt.Println("\tlist: List all available problems and links in db")
		return nil
	} else {
		subcommand := args[0]

		switch subcommand {
			case "add": 
				fmt.Println("add invoked")
			case "delete": 
				fmt.Println("delete invoked")
			case "start": 
				fmt.Println("start invoked")
			case "done": 
				fmt.Println("done invoked")
			case "list": 
				fmt.Println("list invoked")
			default:
				fmt.Println("Unknown problem subcommand")
				return nil 
		}
	}

	return nil
}