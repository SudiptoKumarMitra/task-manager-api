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
func taskHandler( w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "No tasks available")
}
func main() {
	http.HandleFunc("/",homeHandler)
	http.HandleFunc("/health",healthHandler)
	http.HandleFunc("/tasks",taskHandler)
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}