package cmd

import (
	"database/sql"
	"fmt"
)

type DbCommand struct {
	DB *sql.DB
}

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
		fmt.Println("\treset: clears all tables, use -k flag to keep problems in db")
		return nil
	} else {
		subcommand := args[0]

		switch subcommand {
		case "setup":
			fmt.Println("Setting up randomizer.db")
			setupTables(d.DB)
			defaultData(d.DB)
			fmt.Println("Database setup complete.")
		case "reset":
			keepProblemsTable := false

			if len(args) > 1 && args[1] == "-k" {
				keepProblemsTable = true
			}

			if keepProblemsTable {
				fmt.Println("YOU ARE TRYING TO DELETE ALL DATA EXCEPT THE PROBLEMS YOU HAVE ENTERED, TYPE 'yes' TO CONTINUE")
			} else {
				fmt.Println("YOU ARE TRYING TO DELETE ALL DATA, TYPE 'yes' TO CONTINUE")
			}

			var confirmation string
			fmt.Scan(&confirmation)

			if confirmation != "yes" {
				fmt.Println("Aborted reset")
				return nil
			}

			fmt.Println("Delete tables")
			deleteTables(d.DB, keepProblemsTable)

			fmt.Println("Setting up randomizer.db")
			setupTables(d.DB)
			defaultData(d.DB)
			fmt.Println("Database setup complete.")
		default:
			fmt.Println("Unknown db subcommand")
			return nil
		}
	}

	return nil
}

func setupTables(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS problems (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			link TEXT
		);
	`)
	if err != nil {
		fmt.Println("Error creating problems table:", err)
		return
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS assignments (
			id INTEGER PRIMARY KEY,
			problem_id INTEGER REFERENCES problems(id),
			start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			end_time TIMESTAMP
		);
	`)
	if err != nil {
		fmt.Println("Error creating assignments table:", err)
		return
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			timer INTEGER DEFAULT 1,
			streak INTEGER DEFAULT 0,
			last_assignment_id INTEGER REFERENCES assignments(id)
		);
	`)
	if err != nil {
		fmt.Println("Error creating settings table:", err)
		return
	}
}

func defaultData(db *sql.DB) {
	stmt, _ := db.Prepare("INSERT INTO settings (timer, streak) VALUES (?, ?)")
	stmt.Exec(1, 0)
	defer stmt.Close()
}

func deleteTables(db *sql.DB, keepProblemsTable bool) {
	_, err := db.Exec(`
		DROP TABLE settings;
	`)

	if err != nil {
		fmt.Println("Error dropping settings table:", err)
	}

	_, err = db.Exec(`
		DROP TABLE assignments;
	`)

	if err != nil {
		fmt.Println("Error dropping assignments table:", err)
	}

	if !keepProblemsTable {
		_, err = db.Exec(`
			DROP TABLE problems;
		`)

		if err != nil {
			fmt.Println("Error dropping problems table:", err)

		}
	} else {
		fmt.Println("Skipped deleting problems table")
	}
}
