package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func initialTaskModel() ModelTask {
	ti := textarea.New()
	ti.Placeholder = "Add your task here..."
	ti.Focus()

	return ModelTask{
		textarea: ti,
		err:      nil,
	}
}

func (m ModelTask) Init() tea.Cmd {
	return textarea.Blink
}

func (m ModelTask) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m ModelTask) View() string {
	return fmt.Sprintf(
		"What do you need to do next?\n\n%s\n\n%s",
		m.textarea.View(),
		"(ctrl+c to quit)",
	) + "\n\n"
}

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
