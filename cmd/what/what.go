package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/spf13/cobra"
)

// What next - Offer a UI for me to type out what I should be doing next. Stretch goal would be to add the task to a specific project. Find a way to create a project. It could default to the project you're in?
// What now - If you're in a folder it defaults to show the tasks for that folder. Otherwise, it lets you choose which project to look at. You can also check off things that are done.
// Next time you do what now, it won't have that task.

type TaskModel struct {
	ID              int64
	TaskDescription string
	IsComplete      bool
	Project         string
}

type Task struct {
	Description string
	Done        bool
}

type Project struct {
	Name  string
	Tasks []Task
}

type errMsg error
type Model struct {
	Projects []Project
	cursor   int
	selected map[int]struct{}
}

type ModelTask struct {
	textarea textarea.Model
	err      error
}

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

var nextCmd = &cobra.Command{
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

var nowCmd = &cobra.Command{
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

func main() {
	// Check for config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Please set a home directory")
		os.Exit(1)
	}
	_, err = os.Stat(homeDir + "/.what")
	fmt.Println(homeDir)
	if err != nil {
		fmt.Println("Setting up what at ", homeDir)
		err = os.Mkdir(homeDir+"/.what", fs.ModeDir)
		if err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
		os.Create(homeDir + "/.what/what.db")
	}
	cmd, err := initRootCmd()
	if err != nil {
		// TODO: Add Logging
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
