package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

type UserCommand struct {
	DB *sql.DB
}

type SolvedAssignments struct {
	ProblemName string
	StartTime string
	EndTime string
}

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
				if len(args) > 1 {
					hours, err := strconv.Atoi(args[1])
					if err != nil {
						log.Fatal(err)
					}

					if hours < 1 || hours > 24 {
						log.Fatal("Hours must be between 1 and 24 hours")
					}

					u.updateTimer(hours)
				} else {
					fmt.Println("Must pass in number of hours. Usage: ./dsa-randomizer user timer <hours>")
				}
				return nil
			case "streak":
				streak := u.getStreak()
				fmt.Println("Current streak is", streak)
				return nil
			case "history":
				fmt.Println("history was invoked")
			default:
				fmt.Println("Unknown user subcommand")
				return nil
		}
	}

	return nil
}

func (u *UserCommand) updateTimer(hours int) {
	stmt, _ := u.DB.Prepare(`
		UPDATE settings SET timer = ? WHERE id = ?
	`)

	stmt.Exec(hours, 1)
}

func (u *UserCommand) getStreak() int {
	var streak int
	if err := u.DB.QueryRow(`
		SELECT streak FROM settings WHERE id = ?
	`, 1).Scan(&streak); err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("No rows found for streak")
		}
	}

	return streak
}