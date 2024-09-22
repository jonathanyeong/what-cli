package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func initNextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Add a new task",
		Run: func(cmd *cobra.Command, args []string) {
			homeDir, _ := os.UserHomeDir()
			db, _ := sql.Open("sqlite3", homeDir+"/.what/what.db")
			defer db.Close()
			_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tasks (id INTEGER PRIMARY KEY AUTOINCREMENT, taskDescription TEXT, isComplete BOOLEAN, project VARCHAR(255))`)
			if err != nil {
				log.Fatal(err)
			}
			task := initialTaskModel()
			p := tea.NewProgram(task)
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			fmt.Println(task.textarea.Value())
			res, err := db.Exec("INSERT INTO tasks (taskDescription, isComplete, project) VALUES (?, ?, ?)", task.textarea.Value(), false, "Default")
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			id, _ := res.LastInsertId()
			fmt.Println(id)
		},
	}
	return cmd
}
