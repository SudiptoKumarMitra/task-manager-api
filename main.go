package main
import (
	"fmt"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
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

func main() {
	if err:= connectDB(); err!= nil{
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database successfully")	
	r:= gin.Default()
	protected := r.Group("/tasks")
	protected.Use(AuthMiddleware)
	protected.GET("",handleGetTask)
	protected.POST("",handlePostTask)
	// r.POST("/tasks",handlePostTask)
	r.PUT("/tasks/:id",handlePutTask)
	r.DELETE("/tasks/:id",handleDeleteTask)
	fmt.Println("Server is running on port 8080")
	r.Run(":8080")
}