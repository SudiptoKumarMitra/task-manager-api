package main
import (
	"fmt"
	"net/http"
	"log"
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
	if err:= connectDB(); err!= nil{
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database successfully")	
	if err := loadTasks(); err != nil {
		log.Fatal("Failed to load tasks:", err)
	}
	http.HandleFunc("/",homeHandler)
	http.HandleFunc("/health",healthHandler) 
	http.HandleFunc("/tasks",taskHandler)
	http.HandleFunc("/users",usersHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}