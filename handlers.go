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
			for i := range Tasks {
				if Tasks[i].ID == id {
					headerSet(w)
					json.NewEncoder(w).Encode(Tasks[i])
					return
				}
			}
			http.Error(w, "Task Not Found", http.StatusNotFound)
		} else {
			headerSet(w)
			json.NewEncoder(w).Encode(Tasks)
		}
}

func handlePostTask(w http.ResponseWriter, r *http.Request) {
	var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil{
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		task.ID = nextID
		nextID++
		Tasks=append(Tasks,task)
		if err := saveTasks(); err != nil {
			http.Error(w, "Failed to save JSON", http.StatusInternalServerError)
			return
		}
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