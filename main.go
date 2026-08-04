package main
import (
	"fmt"
	"net/http"
)
func homeHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Welcome to Task Manager API")
}
func healthHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Server is healthy")
}
func usersHandle(w http.ResponseWriter, r *http.Request){
	switch r.Method {
		case http.MethodGet:
			fmt.Fprintln(w,"Users List")
		case http.MethodPost:
			fmt.Fprintln(w,"User Created")
		case http.MethodPut:
			fmt.Fprintln(w,"User Updated")
		case http.MethodDelete:
			fmt.Fprintln(w,"User Deleted")
		default:
			http.Error(w,"Method Not Allowed",http.StatusMethodNotAllowed)
	}
}
func taskHandler( w http.ResponseWriter, r *http.Request){
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintln(w, "Task List")
	case http.MethodPost:
		fmt.Fprintln(w, "Task Created")
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
	http.HandleFunc("/users",usersHandle)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}