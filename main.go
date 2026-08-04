package main
import (
	"fmt"
	"net/http"
	"encoding/json"
	"strconv"
)
type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
}
type Task struct {
	ID int `json:"id"`
	Title string `json:"title"`
}
var Tasks = []Task{}
var nextID = 1
func homeHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Welcome to Task Manager API")
}
func healthHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Server is healthy")
}
func usersHandler(w http.ResponseWriter, r *http.Request){
	
	switch r.Method {
		case http.MethodGet:
			fmt.Fprintln(w, "User List")
		case http.MethodPost:
			fmt.Fprintln(w, "User Created")
		case http.MethodPut:
			fmt.Fprintln(w, "User Updated")
		case http.MethodDelete:
			fmt.Fprintln(w, "User Deleted")
		default:
			http.Error(w,"Method Not Allowed",http.StatusMethodNotAllowed)
	}
}
func taskHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case http.MethodGet:
		if r.URL.Query().Get("id") != "" {
			id, err := strconv.Atoi(r.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			if id > len(Tasks) || id < 1 {
				http.Error(w, "Task Not Found", http.StatusNotFound)
				return
			}
			for _, task := range Tasks {
				if task.ID == id {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(task)
					return
				}
			}
		} else {
			w.Header().Set("Content-type", "application/json")
			json.NewEncoder(w).Encode(Tasks)
		}
	case http.MethodPost:
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil{
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		task.ID = nextID
		nextID++
		Tasks=append(Tasks,task)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	case http.MethodPut:
		fmt.Fprintln(w, "Task Updated")
	case http.MethodDelete:
		fmt.Fprintln(w, "Task Deleted")
	default:
		http.Error(w,"Method Not allowed", http.StatusMethodNotAllowed)
	}
}
func main() {
	http.HandleFunc("/",homeHandler)
	http.HandleFunc("/health",healthHandler) 
	http.HandleFunc("/tasks",taskHandler)
	http.HandleFunc("/users",usersHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}