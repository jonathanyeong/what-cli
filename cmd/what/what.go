package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
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
