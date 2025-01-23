package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type UserCommand struct {
	DB *sql.DB
}

type Assignment struct {
	ProblemName string
	ProblemLink string
	StartTime   sql.NullTime
	EndTime     sql.NullTime
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

				updateTimer(u.DB, hours)
			} else {
				fmt.Println("Must pass in number of hours. Usage: ./dsa-randomizer user timer <hours>")
			}
			return nil
		case "streak":
			streak := getStreak(u.DB)
			fmt.Println("Current streak is", streak)
			return nil
		case "history":
			getHistory(u.DB)
		default:
			fmt.Println("Unknown user subcommand")
			return nil
		}
	}

	return nil
}

func updateTimer(db *sql.DB, hours int) {
	stmt, _ := db.Prepare(`
		UPDATE settings SET timer = ? WHERE id = ?
	`)

	stmt.Exec(hours, 1)
}

func getStreak(db *sql.DB) int {
	var streak int
	if err := db.QueryRow(`
		SELECT streak FROM settings WHERE id = ?
	`, 1).Scan(&streak); err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("No rows found for streak")
		}
	}

	return streak
}

func getHistory(db *sql.DB) {
	assignments := []Assignment{}

	rows, err := db.Query(`
		SELECT p.name, p.link, a.start_time, a.end_time 
		FROM assignments a
		JOIN problems p ON a.problem_id = p.id
	`)
	if err != nil {
		log.Fatalf("Error querying database: %v", err)
	}

	defer rows.Close()

	for rows.Next() {
		var assignment Assignment
		if err := rows.Scan(&assignment.ProblemName, &assignment.ProblemLink, &assignment.StartTime, &assignment.EndTime); err != nil {
			log.Fatal(err)
		}
		assignments = append(assignments, assignment)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Printf("%-50s %-35s %-35s\n", "Problem Name", "Start Time", "End Time")
	fmt.Println(strings.Repeat("-", 100))

	for _, assignment := range assignments {
		fmt.Printf("%-50s %-35v %-35v\n",
			assignment.ProblemName,
			formatNullTime(assignment.StartTime),
			formatNullTime(assignment.EndTime),
		)
	}
}

func formatNullTime(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format(time.RFC1123)
	}

	return "No time set"
}
