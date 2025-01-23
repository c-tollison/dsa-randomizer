package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/dsa-randomizer/cmd"
	_ "github.com/mattn/go-sqlite3"
)

const db_name string = "randomizer.db"

func help(commands []cmd.Command) {
	fmt.Println("Usage: dsa-randomizer <command>")
	fmt.Println("Commands:")
	for _, c := range commands {
		fmt.Println("\t", c.Command() + ":", c.Help())
	}
}

func main() {
	args := os.Args[1:]
	db, err := sql.Open("sqlite3", path.Base(db_name))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	commands := []cmd.Command{
		&cmd.DbCommand{DB: db},
		&cmd.UserCommand{DB: db},
		&cmd.ProblemCommand{DB: db},
	}

	if len(args) == 0 || args[0] == "help" {
		help(commands)
	} else {
		found := false

		for _, c := range commands {
			if c.Command() == args[0] {
				found = true
				c.Run(args[1:])
			}
		}

		if !found {
			help(commands)
		}
	}
}