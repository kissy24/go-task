package main

import (
	"fmt"
	"os"
	"strings"

	"zan/internal/app"
	"zan/internal/task"

	"github.com/spf13/cobra"
)

var (
	appInstance *app.App
)

func init() {
	var err error
	appInstance, err = app.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing app: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "zan",
		Short: "Zan is a lightweight and fast CLI task management system",
		Long:  `Zan is a lightweight and fast CLI task management system built with Go and Bubble Tea.`,
		Run: func(cmd *cobra.Command, args []string) {
			// デフォルトの動作（引数なしで実行された場合）はリスト表示
			listTasks(cmd, args)
		},
	}

	// Add command
	addCmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run:   addTask,
	}
	addCmd.Flags().StringP("description", "d", "", "Detailed description of the task")
	addCmd.Flags().StringP("priority", "p", "", "Priority of the task (High, Medium, Low)")
	addCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags for the task")
	rootCmd.AddCommand(addCmd)

	// List command
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Run:   listTasks,
	}
	rootCmd.AddCommand(listCmd)

	// Show command
	showCmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show task details",
		Args:  cobra.ExactArgs(1),
		Run:   showTask,
	}
	rootCmd.AddCommand(showCmd)

	// Complete command
	completeCmd := &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a task as complete",
		Args:  cobra.ExactArgs(1),
		Run:   completeTask,
	}
	rootCmd.AddCommand(completeCmd)

	// Delete command
	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		Run:   deleteTask,
	}
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func addTask(cmd *cobra.Command, args []string) {
	title := args[0]
	description, _ := cmd.Flags().GetString("description")
	priorityStr, _ := cmd.Flags().GetString("priority")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	priority := task.Priority(strings.ToUpper(priorityStr))

	t, err := appInstance.AddTask(title, description, priority, tags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding task: %v\n", err)
		return
	}
	fmt.Printf("Task added: %s - %s\n", t.ID, t.Title)
}

func listTasks(cmd *cobra.Command, args []string) {
	tasks := appInstance.GetAllTasks()
	total, completed, incomplete := appInstance.GetTaskStats()

	if len(tasks) == 0 {
		fmt.Println("No tasks found. Add a new task using 'zan add <title>'")
		return
	}

	fmt.Println("ID         Status    Priority  Title")
	fmt.Println("--------------------------------------------------")
	for _, t := range tasks {
		statusIcon := getStatusIcon(t.Status)
		priorityColor := getPriorityColor(t.Priority)
		fmt.Printf("%-10s %-9s %s%-9s\033[0m %s\n", t.ID[:8], statusIcon, priorityColor, t.Priority, t.Title)
	}
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Total: %d | Incomplete: %d | Completed: %d\n", total, incomplete, completed)
}

func showTask(cmd *cobra.Command, args []string) {
	id := args[0]
	t, err := appInstance.GetTaskByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error showing task: %v\n", err)
		return
	}

	fmt.Println("Task Details:")
	fmt.Printf("  ID:          %s\n", t.ID)
	fmt.Printf("  Title:       %s\n", t.Title)
	fmt.Printf("  Description: %s\n", t.Description)
	fmt.Printf("  Status:      %s %s\n", getStatusIcon(t.Status), t.Status)
	fmt.Printf("  Priority:    %s%s\033[0m %s\n", getPriorityColor(t.Priority), t.Priority, t.Priority)
	fmt.Printf("  Tags:        %s\n", strings.Join(t.Tags, ", "))
	fmt.Printf("  Created At:  %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Updated At:  %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))
	if t.CompletedAt != nil {
		fmt.Printf("  Completed At: %s\n", t.CompletedAt.Format("2006-01-02 15:04:05"))
	}
}

func completeTask(cmd *cobra.Command, args []string) {
	id := args[0]
	_, err := appInstance.UpdateTask(id, "", "", task.StatusDone, "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
		return
	}
	fmt.Printf("Task %s marked as DONE.\n", id)
}

func deleteTask(cmd *cobra.Command, args []string) {
	id := args[0]
	fmt.Printf("Are you sure you want to delete task %s? This action cannot be undone. (yes/no): ", id)
	var confirmation string
	fmt.Scanln(&confirmation)

	if strings.ToLower(confirmation) != "yes" {
		fmt.Println("Task deletion cancelled.")
		return
	}

	err := appInstance.DeleteTask(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
		return
	}
	fmt.Printf("Task %s deleted.\n", id)
}

func getStatusIcon(status task.Status) string {
	switch status {
	case task.StatusTODO:
		return "●"
	case task.StatusInProgress:
		return "◐"
	case task.StatusDone:
		return "✓"
	case task.StatusPending:
		return "⏸"
	default:
		return "?"
	}
}

func getPriorityColor(priority task.Priority) string {
	// ANSI escape codes for colors
	const (
		Red    = "\033[31m"
		Yellow = "\033[33m"
		Green  = "\033[32m"
		Reset  = "\033[0m"
	)
	switch priority {
	case task.PriorityHigh:
		return Red
	case task.PriorityMedium:
		return Yellow
	case task.PriorityLow:
		return Green
	default:
		return Reset
	}
}
