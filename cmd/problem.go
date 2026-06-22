package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"math"
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
		fmt.Println("\tlist: list all problems with spaced repetition status")
		fmt.Println("\tstart: start a due problem (or random if none due)")
		fmt.Println("\tcurrent: provides current problem")
		fmt.Println("\tdone: marks current problem complete and updates spaced repetition schedule")
		fmt.Println("\treview: show today's due count and upcoming reviews")
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
		case "review":
			reviewStatus(p.DB)
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
	rows, err := db.Query(`
		SELECT id, name, link, COALESCE(repetitions, 0),
		       COALESCE(strftime('%Y-%m-%d', next_review), 'New')
		FROM problems
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type row struct {
		id          int
		name        string
		link        string
		repetitions int
		nextReview  string
	}

	var problems []row
	for rows.Next() {
		var p row
		if err := rows.Scan(&p.id, &p.name, &p.link, &p.repetitions, &p.nextReview); err != nil {
			log.Fatal(err)
		}
		problems = append(problems, p)
	}

	fmt.Printf("%-10s %-40s %-60s %-14s %-10s\n", "ID", "Problem Name", "Problem Link", "Next Review", "Reviews")
	fmt.Println(strings.Repeat("-", 140))
	for _, p := range problems {
		fmt.Printf("%-10d %-40s %-60s %-14s %-10d\n", p.id, p.name, p.link, p.nextReview, p.repetitions)
	}
}

// sm2Update applies the SM-2 spaced repetition algorithm and returns the updated
// interval (days), ease factor, and repetition count.
// quality: 0-2 = failed, 3 = hard, 4 = medium, 5 = easy
func sm2Update(interval int, easeFactor float64, repetitions int, quality int) (int, float64, int) {
	if quality < 3 {
		repetitions = 0
		interval = 1
	} else {
		switch repetitions {
		case 0:
			interval = 1
		case 1:
			interval = 6
		default:
			interval = int(math.Round(float64(interval) * easeFactor))
		}
		repetitions++
	}
	ef := easeFactor + 0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02)
	if ef < 1.3 {
		ef = 1.3
	}
	return interval, ef, repetitions
}

func startProblem(db *sql.DB) {
	var dueCount int
	db.QueryRow(`
		SELECT COUNT(*) FROM problems
		WHERE next_review IS NULL OR date(next_review) <= date('now')
	`).Scan(&dueCount)

	var row Problem
	var isDue bool

	dueErr := db.QueryRow(`
		SELECT id, name, link FROM problems
		WHERE next_review IS NULL OR date(next_review) <= date('now')
		ORDER BY RANDOM() LIMIT 1
	`).Scan(&row.id, &row.name, &row.link)

	if dueErr == sql.ErrNoRows {
		if err := db.QueryRow(`
			SELECT id, name, link FROM problems ORDER BY RANDOM() LIMIT 1
		`).Scan(&row.id, &row.name, &row.link); err != nil {
			if err == sql.ErrNoRows {
				log.Fatal("No rows found in problems table")
			}
			log.Fatal(err)
		}
	} else if dueErr != nil {
		log.Fatal(dueErr)
	} else {
		isDue = true
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

	if isDue {
		fmt.Printf("%d problem(s) due for review today.\n", dueCount)
	} else {
		fmt.Println("No problems due for review. Here is a random problem:")
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
		UPDATE assignments SET end_time = datetime('now', 'utc')
		WHERE id IN (SELECT last_assignment_id FROM settings WHERE id = ?)
	`, 1)

	var startTime, endTime time.Time
	if err = db.QueryRow(`
		SELECT start_time, end_time
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

	_, err = db.Exec(`UPDATE settings SET streak = ? WHERE id = ?`, streak, 1)

	// Spaced repetition: prompt for recall quality and schedule next review.
	fmt.Println("\nRate your recall:")
	fmt.Println("  1: Easy   - recalled without effort")
	fmt.Println("  2: Medium - recalled with some effort")
	fmt.Println("  3: Hard   - struggled significantly or forgot")

	var ratingInput int
	fmt.Scan(&ratingInput)

	qualityMap := map[int]int{1: 5, 2: 3, 3: 1}
	quality, ok := qualityMap[ratingInput]
	if !ok {
		fmt.Println("Invalid rating, defaulting to Medium")
		quality = 3
	}

	var problemId, interval, repetitions int
	var easeFactor float64
	if err = db.QueryRow(`
		SELECT p.id, COALESCE(p.interval, 1), COALESCE(p.ease_factor, 2.5), COALESCE(p.repetitions, 0)
		FROM assignments a
		JOIN problems p ON a.problem_id = p.id
		WHERE a.id = (SELECT last_assignment_id FROM settings WHERE id = 1)
	`).Scan(&problemId, &interval, &easeFactor, &repetitions); err != nil {
		log.Fatal(err)
	}

	newInterval, newEF, newRep := sm2Update(interval, easeFactor, repetitions, quality)
	nextReview := time.Now().AddDate(0, 0, newInterval).Format("2006-01-02")

	if _, err = db.Exec(`
		UPDATE problems SET interval = ?, ease_factor = ?, repetitions = ?, next_review = ?
		WHERE id = ?
	`, newInterval, newEF, newRep, nextReview, problemId); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Next review scheduled in %d day(s) on %s\n", newInterval, nextReview)
}

func reviewStatus(db *sql.DB) {
	var dueCount int
	if err := db.QueryRow(`
		SELECT COUNT(*) FROM problems
		WHERE next_review IS NULL OR date(next_review) <= date('now')
	`).Scan(&dueCount); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d problem(s) due for review today\n", dueCount)

	rows, err := db.Query(`
		SELECT name, strftime('%Y-%m-%d', next_review), COALESCE(repetitions, 0)
		FROM problems
		WHERE date(next_review) > date('now')
		AND date(next_review) <= date('now', '+7 days')
		ORDER BY next_review
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	type upcoming struct {
		name       string
		nextReview string
		reps       int
	}
	var items []upcoming
	for rows.Next() {
		var u upcoming
		rows.Scan(&u.name, &u.nextReview, &u.reps)
		items = append(items, u)
	}

	if len(items) > 0 {
		fmt.Println("\nDue in the next 7 days:")
		fmt.Printf("%-50s %-14s %-12s\n", "Problem", "Next Review", "Reviews Done")
		fmt.Println(strings.Repeat("-", 80))
		for _, u := range items {
			fmt.Printf("%-50s %-14s %-12d\n", u.name, u.nextReview, u.reps)
		}
	}
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
