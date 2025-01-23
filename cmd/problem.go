package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type ProblemCommand struct {
	DB *sql.DB
}

type Problem struct {
	id   int
	name string
	link string
}

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
		fmt.Println("\tadd: adds problem to the db, usage: dsa-randomizer problem add <name> <link>")
		fmt.Println("\tdelete: deletes problem from the db, usage: dsa-randomizer problem delete <id>")
		fmt.Println("\tlist: List all available problems and links in db")
		fmt.Println("\tstart: provides a random problem and starts a new timer")
		fmt.Println("\tcurrent: provides current problem")
		fmt.Println("\tdone: marks current open problem as complete, boosts streak if within timer set on the user")
		return nil
	} else {
		subcommand := args[0]

		switch subcommand {
		case "add":

			if len(args) == 3 {
				name := args[1]
				link := args[2]

				addProblem(p.DB, name, link)
			} else {
				fmt.Println("Usage: dsa-randomizer problem add <name> <link>")
				return nil
			}
		case "delete":
			if len(args) == 2 {
				id, err := strconv.Atoi(args[1])
				if err != nil {
					log.Fatal(err)
				}

				deleteProblem(p.DB, id)
			} else {
				fmt.Println("Usage: dsa-randomizer problem delete <id>")
			}
		case "list":
			listProblems(p.DB)
		case "start":
			startProblem(p.DB)
		case "current":
			currentProblem(p.DB)
		case "done":
			completeProblem(p.DB)
		default:
			fmt.Println("Unknown problem subcommand")
			return nil
		}
	}

	return nil
}

func addProblem(db *sql.DB, problemName string, problemLink string) {
	stmt, _ := db.Prepare("INSERT INTO problems (name, link) values (?, ?)")
	stmt.Exec(problemName, problemLink)
	defer stmt.Close()
}

func deleteProblem(db *sql.DB, id int) {
	stmt, _ := db.Prepare("DELETE FROM problems WHERE id = ?")
	stmt.Exec(id)
	defer stmt.Close()
}

func listProblems(db *sql.DB) {
	problems := []Problem{}
	rows, _ := db.Query("SELECT * FROM problems")
	defer rows.Close()

	for rows.Next() {
		var problem Problem
		if err := rows.Scan(&problem.id, &problem.name, &problem.link); err != nil {
			log.Fatal(err)
		}
		problems = append(problems, problem)
	}

	fmt.Printf("%-30s %-50s %-80s\n", "Problem ID", "Problem Name", "Problem Link")
	fmt.Println(strings.Repeat("-", 160))
	for _, p := range problems {
		fmt.Printf("%-30d %-50s %-80s\n", p.id, p.name, p.link)
	}
}

func startProblem(db *sql.DB) {
	var row Problem
	if err := db.QueryRow(`
		SELECT * FROM problems ORDER BY RANDOM() LIMIT 1
	`).Scan(&row.id, &row.name, &row.link); err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("No rows found in problems table")
		} else {
			log.Fatal(err)
		}
	}

	var hours int
	var lastAssignmentId sql.NullInt64

	if err := db.QueryRow(`
		SELECT timer, last_assignment_id FROM settings WHERE id = ?
	`, 1).Scan(&hours, &lastAssignmentId); err != nil {
		log.Fatal(err)
	}

	if lastAssignmentId.Valid {
		var endTimeExists bool
		if err := db.QueryRow(`
			SELECT end_time IS NOT NULL 
			FROM assignments 
			WHERE id = ?
    	`, lastAssignmentId).Scan(&endTimeExists); err != nil {
			log.Fatal(err)
		}

		if !endTimeExists {
			fmt.Println("There is already an active problem")
			return
		}
	}

	var assignmentId int
	if err := db.QueryRow(`
		INSERT INTO assignments (problem_id, start_time) 
		VALUES (?, datetime('now', 'utc')) 
		RETURNING id
	`, row.id).Scan(&assignmentId); err != nil {
		log.Fatal(err)
	}

	_, err := db.Exec(`
		UPDATE settings SET last_assignment_id = ? WHERE id = ?
	`, assignmentId, 1)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting problem:", row.name)
	fmt.Println("Here is the link:", row.link)
	fmt.Print("You have ", hours, " hour")
	if hours > 1 {
		fmt.Print("s")
	}
	fmt.Println()

}

func completeProblem(db *sql.DB) {
	var hours, streak int
	if err := db.QueryRow(`
		SELECT timer, streak FROM settings WHERE id = ?
	`, 1).Scan(&hours, &streak); err != nil {
		log.Fatal(err)
	}

	var endTimeExists bool
	err := db.QueryRow(`
        SELECT end_time IS NOT NULL 
        FROM assignments 
        WHERE id IN (SELECT last_assignment_id FROM settings WHERE id = ?)
    `, 1).Scan(&endTimeExists)

	if err != nil {
		log.Fatal(err)
	}

	if endTimeExists {
		fmt.Println("Assignment already completed")
		return
	}

	_, err = db.Exec(`
		UPDATE assignments SET end_time = datetime('now', 'utc') WHERE id in 
		(SELECT last_assignment_id FROM settings WHERE id = ?)
	`, 1)

	var startTime, endTime time.Time

	if err = db.QueryRow(`
		SELECT 
			start_time, 
			end_time 
		FROM assignments 
		WHERE id IN (SELECT last_assignment_id FROM settings WHERE id = ?)
	`, 1).Scan(&startTime, &endTime); err != nil {
		log.Fatal(err)
	}

	duration := endTime.Sub(startTime)
	withinThreshold := duration <= time.Duration(hours)*time.Hour

	if withinThreshold {
		streak++
		fmt.Println("Completed within time frame, streak is now", streak)
	} else {
		streak = 0
		fmt.Println("Not completed within time frame, streak is now", streak)
	}

	_, err = db.Exec(`
		UPDATE settings SET streak = ? WHERE id = ?
	`, streak, 1)
}

func currentProblem(db *sql.DB) {
	var lastAssignmentId sql.NullInt64

	if err := db.QueryRow(`
		SELECT last_assignment_id FROM settings WHERE id = ?
	`, 1).Scan(&lastAssignmentId); err != nil {
		log.Fatal(err)
	}

	if lastAssignmentId.Valid {
		var assignment Assignment
		if err := db.QueryRow(`
			SELECT p.name, p.link, a.start_time, a.end_time 
			FROM assignments a
			JOIN problems p ON a.problem_id = p.id
			WHERE a.id = ?
    	`, lastAssignmentId).Scan(&assignment.ProblemName, &assignment.ProblemLink, &assignment.StartTime, &assignment.EndTime); err != nil {
			log.Fatal(err)
		}

		if !assignment.EndTime.Valid {
			fmt.Printf("%-50s %-80s %-35s\n", "Problem Name", "Problem Link", "Start Time")
			fmt.Println(strings.Repeat("-", 160))
			fmt.Printf("%-50s %-80s %-35v\n",
				assignment.ProblemName,
				assignment.ProblemLink,
				formatNullTime(assignment.StartTime),
			)
			return
		} else {
			fmt.Println("There is no current problem")
			return
		}
	} else {
		fmt.Println("There is no current problem")
		return
	}
}
