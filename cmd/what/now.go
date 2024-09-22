package main

import (
	"database/sql"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func initialModel(tasks []TaskModel) Model {
	var taskTransform []Task

	for _, task := range tasks {
		taskTransform = append(taskTransform, Task{Description: task.TaskDescription, Done: task.IsComplete})
	}
	fmt.Printf("Tasks found: %v\n", tasks)
	fmt.Printf("taskTransform found: %v\n", taskTransform)

	return Model{
		Projects: []Project{
			{Name: "Default", Tasks: taskTransform},
		},
		selected: make(map[int]struct{}),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Projects[0].Tasks)-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := "What Now?\n\n"
	for i, task := range m.Projects[0].Tasks {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, task.Description)
	}
	s += "\nPress q to quit.\n"
	return s
}

func initNowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "now",
		Short: "View and manage current tasks",
		Run: func(cmd *cobra.Command, args []string) {
			var tasks []TaskModel
			homeDir, _ := os.UserHomeDir()
			db, _ := sql.Open("sqlite3", homeDir+"/.what/what.db")
			defer db.Close()
			rows, err := db.Query("SELECT * FROM tasks")
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			defer rows.Close()

			for rows.Next() {
				var task TaskModel
				if err := rows.Scan(&task.ID, &task.TaskDescription, &task.IsComplete, &task.Project); err != nil {
					fmt.Printf("Error: %v", err)
				}
				tasks = append(tasks, task)
			}

			p := tea.NewProgram(initialModel(tasks))
			if _, err := p.Run(); err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
		},
	}
	return cmd
}
