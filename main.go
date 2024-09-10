package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// What next - Offer a UI for me to type out what I should be doing next. Stretch goal would be to add the task to a specific project. Find a way to create a project. It could default to the project you're in?
// What now - If you're in a folder it defaults to show the tasks for that folder. Otherwise, it lets you choose which project to look at. You can also check off things that are done.
// Next time you do what now, it won't have that task.

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

func initialModel() Model {
	return Model{
		Projects: []Project{
			{Name: "Default", Tasks: []Task{{Description: "sample task", Done: false}}},
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

var rootCmd = &cobra.Command{
	Use:   "what",
	Short: "A task manager inside your CLI!",
}

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Add a new task",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialTaskModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	},
}

var nowCmd = &cobra.Command{
	Use:   "now",
	Short: "View and manage current tasks",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	},
}

func main() {
	rootCmd.AddCommand(nextCmd)
	rootCmd.AddCommand(nowCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
