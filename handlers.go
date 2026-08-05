package main
import (
	"encoding/json"
	"net/http"
)
func handleGetTask(w http.ResponseWriter, r *http.Request) {
if r.URL.Query().Get("id") != "" {
			id, err := getID(r)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			var task Task
			err = DB.QueryRow("SELECT id,title FROM tasks WHERE id = $1",id).Scan(&task.ID,&task.Title)
			if err != nil {
				http.Error(w, "Task Not Found", http.StatusNotFound)
				return
			}
			headerSet(w)
			json.NewEncoder(w).Encode(task)
		} else {
			rows,err := DB.Query("SELECT id,title FROM tasks")
			if err != nil {
				http.Error(w, "Failed to connect to Database", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			var tasks []Task
			for rows.Next() {
				var task Task
				err := rows.Scan(&task.ID, &task.Title)
					if err != nil {
					http.Error(w, "Error in fetching data", http.StatusInternalServerError)
					return
				}
				tasks=append(tasks,task)
			}
			headerSet(w)
			json.NewEncoder(w).Encode(tasks)
		}
}

func handlePostTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil{
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var id int
	err = DB.QueryRow("INSERT INTO tasks (title) VALUES ($1) RETURNING id",task.Title).Scan(&id)
	if err != nil {
		http.Error(w, "Failed to Insert into database",http.StatusInternalServerError)
		return
	}
	task.ID = id
	headerSet(w)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func handlePutTask(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("id") != "" {
			id, err := getID(r)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			var task Task
			err = json.NewDecoder(r.Body).Decode(&task)
				if err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}
			for i := range Tasks {
				if Tasks[i].ID == id {
				Tasks[i].Title = task.Title
				headerSet(w)
				json.NewEncoder(w).Encode(Tasks[i])
				if err := saveTasks(); err != nil {
					http.Error(w, "Failed to save JSON", http.StatusInternalServerError)
					return
				}
				return
				}
			}
			http.Error(w, "Task Not Found", http.StatusNotFound)
		}
}
func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("id") != "" {
			id, err := getID(r)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			for i := range Tasks{
				if Tasks[i].ID == id {
					Tasks=append(Tasks[:i],Tasks[i+1:]...)
					if err := saveTasks(); err != nil {
						http.Error(w, "Failed to save JSON", http.StatusInternalServerError)
						return
					}
					headerSet(w)
					json.NewEncoder(w).Encode("Task Deleted")
					return
				}
			}
			http.Error(w, "Task Not Found", http.StatusNotFound)
		}
}