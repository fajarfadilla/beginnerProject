package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

type TaskList struct {
	Tasks []Task `json:"tasks"`
}

func check(err error, msg string) {
	if err != nil {
		log.Printf("%s: %v", msg, err)
	}
}
func initializeTaskFile(filename string) error {
	taskList := TaskList{Tasks: []Task{}}
	data, err := json.MarshalIndent(taskList, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling empty task list: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing initial task file: %w", err)
	}
	return nil
}

// loadTasks reads and decodes tasks from the JSON file
func loadTasks(filename string) (TaskList, error) {
	var taskList TaskList

	data, err := os.ReadFile(filename)
	if err != nil {
		return taskList, fmt.Errorf("error reading file: %w", err)
	}

	if len(data) == 0 {
		return TaskList{Tasks: []Task{}}, nil
	}

	err = json.Unmarshal(data, &taskList)
	if err != nil {
		return taskList, fmt.Errorf("error decoding JSON: %w", err)
	}

	return taskList, nil
}

const FILENAME = "task.json"

// saveTasks writes the task list back to the JSON file
func saveTasks(filename string, taskList *TaskList) error {
	data, err := json.MarshalIndent(taskList, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling task list: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing tasks to file: %w", err)
	}

	return nil
}

func createTask(taskList *TaskList, description string) (id int) {
	if description == "" {
		fmt.Println("task description cannot be empty")
		return
	}
	newTask := Task{}
	if len(taskList.Tasks) < 1 {
		newTask = Task{
			ID:          1,
			Description: description,
			Status:      "TODO",
			CreatedAt:   time.Now().UTC(),
		}
		taskList.Tasks = append(taskList.Tasks, newTask)
		return newTask.ID
	}
	task := taskList.Tasks[len(taskList.Tasks)-1]
	newTask = Task{
		ID:          task.ID + 1,
		Description: description,
		Status:      "TODO",
		CreatedAt:   time.Now().UTC(),
	}
	taskList.Tasks = append(taskList.Tasks, newTask)

	return newTask.ID
}
func updateTask(taskList *TaskList, id int, description string) error {
	if description == "" {
		return fmt.Errorf("task description cannot be empty")
	}
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == id {
			taskList.Tasks[i].Description = description
			taskList.Tasks[i].UpdatedAt = time.Now().UTC()
			return nil
		}
	}
	return fmt.Errorf("no task found with id: %d", id)
}
func updateTaskProgres(taskList *TaskList, id int, command string) error {
	for i := range taskList.Tasks {
		taskID := taskList.Tasks[i].ID
		if taskID == id {
			status := strings.ToLower(taskList.Tasks[i].Status)

			switch command {
			case "mark-in-progress":
				if status == "todo" {
					taskList.Tasks[i].Status = "in-progress"
					taskList.Tasks[i].UpdatedAt = time.Now().UTC()
					return nil
				} else {
					return fmt.Errorf("cannot mark task with status %s to in-progress", status)
				}

			case "mark-done":
				if status == "todo" || status == "in-progress" {
					taskList.Tasks[i].Status = "done"
					taskList.Tasks[i].UpdatedAt = time.Now().UTC()
					return nil
				} else {
					return fmt.Errorf("cannot mark task with status %s to done", status)
				}

			default:
				return fmt.Errorf("unknown command: %s", command)
			}
		}
	}

	return fmt.Errorf("no task found with id: %d", id)
}

func deleteTask(taskList *TaskList, id int) error {
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == id {
			// Remove the task while preserving order
			taskList.Tasks = append(taskList.Tasks[:i], taskList.Tasks[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("no task found with id: %d", id)
}

func listByStatus(status string) {
	tasklist, err := loadTasks(FILENAME)
	check(err, "failed load task")

	found := false
	for _, task := range tasklist.Tasks {
		if strings.ToLower(task.Status) == strings.ToLower(status) {
			out, err := json.MarshalIndent(task, "", " ")
			check(err, "error marshalling task")
			fmt.Println(string(out))
			found = true
		}
	}

	if !found {
		fmt.Printf("No tasks found with status %s\n", status)
	}
}

func ensureFileExists() error {
	_, err := os.Stat(FILENAME)
	if err != nil {
		if os.IsNotExist(err) {
			return initializeTaskFile(FILENAME)
		}
		return err
	}
	return nil
}

func main() {
	input := os.Args
	if len(input) <= 1 {
		fmt.Println("Usage: tasklist <command>")
		fmt.Println("Commands:")
		fmt.Println("  add <task>")
		fmt.Println("  list")
		fmt.Println("  mark-in-progress <task id>")
		fmt.Println("  mark-done <task id>")
		os.Exit(0)
	}

	command := strings.ToLower(os.Args[1])
	switch command {
	case "add":
		if len(input) < 3 {
			fmt.Println("You dont provide description")
			os.Exit(1)
		}
		description := input[2]
		if description == "" {
			fmt.Println("description cannot be empty")
			os.Exit(1)
		}
		if err := ensureFileExists(); err != nil {
			fmt.Printf("Failed initialized file %v", err)
			os.Exit(1)
		}
		tasklist, err := loadTasks(FILENAME)
		check(err, "failed load task")
		taskID := createTask(&tasklist, description)
		err = saveTasks(FILENAME, &tasklist)
		check(err, "failed to save task to file")
		fmt.Println(taskID)
	case "list":
		if len(input) < 3 {
			tasklist, err := loadTasks(FILENAME)
			check(err, "failed load task")
			for _, task := range tasklist.Tasks {
				out, err := json.MarshalIndent(task, "", " ")
				check(err, "error marshalling task")
				fmt.Println(string(out))
			}
		} else {
			listByStatus(input[2])
		}

	case "update":
		if len(input) < 3 {
			fmt.Println("You dont give an id and description")
			os.Exit(1)

		}
		if len(input) < 4 {
			fmt.Println("You dont provide description")
			os.Exit(1)
		}
		taskList, err := loadTasks(FILENAME)
		check(err, "failed load task")
		id, err := strconv.Atoi(input[2])
		description := input[3]
		check(err, "failed to parse id")
		err = updateTask(&taskList, id, description)
		check(err, "failed to update task")
		err = saveTasks(FILENAME, &taskList)
		check(err, "failed to save task to file")
	case "delete":
		if len(input) < 3 {
			fmt.Println("You dont provide id ")
			os.Exit(1)
		}
		taskList, err := loadTasks(FILENAME)
		check(err, "failed load task")
		id, err := strconv.Atoi(input[2])
		err = deleteTask(&taskList, id)
		check(err, "failed delete task")
		err = saveTasks(FILENAME, &taskList)
		check(err, "failed to save task to file")
	case "mark-in-progress":
		if len(input) < 3 {
			fmt.Println("You dont provide id ")
			os.Exit(1)
		}
		taskList, err := loadTasks(FILENAME)
		check(err, "failed load task")
		id, err := strconv.Atoi(input[2])
		command := input[1]
		err = updateTaskProgres(&taskList, id, command)
		check(err, "failed mark status to in-progress")
		err = saveTasks(FILENAME, &taskList)
		check(err, "failed to save task to file")
	case "mark-done":
		if len(input) < 3 {
			fmt.Println("You dont provide id ")
			os.Exit(1)
		}
		taskList, err := loadTasks(FILENAME)
		check(err, "failed load task")
		id, err := strconv.Atoi(input[2])
		command := input[1]
		err = updateTaskProgres(&taskList, id, command)
		check(err, "failed mark status to in-progress")
		err = saveTasks(FILENAME, &taskList)
		check(err, "failed to save task to file")
	}
}
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}
