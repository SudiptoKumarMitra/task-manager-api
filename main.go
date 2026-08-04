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

func getID(r *http.Request) (int,error){
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	return id,err
}

func headerSet(w http.ResponseWriter){
	w.Header().Set("Content-Type", "application/json")
}
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
					headerSet(w)
					json.NewEncoder(w).Encode("Task Deleted")
					return
				}
			}
			http.Error(w, "Task Not Found", http.StatusNotFound)
		}
}
func taskHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case http.MethodGet:
		handleGetTask(w, r)
	case http.MethodPost:
		handlePostTask(w, r)
	case http.MethodPut:
		handlePutTask(w, r)
	case http.MethodDelete:
		handleDeleteTask(w, r)
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