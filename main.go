package main
import (
	"fmt"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
	"task-manager-api/repository"
	"task-manager-api/service"
)
type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
}
type RegisterRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}
type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}
type Config struct {
	JWT_SECRET string 
	PORT string 
	DBHost string
	DBPort string
	DBUser string
	DBPassword string
	DBName string
}
func loadConfig() Config {
	return Config{
		JWT_SECRET: os.Getenv("JWT_SECRET"),
		PORT: os.Getenv("PORT"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBUser: os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"),
	}
}

var config Config
func homeHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Welcome to Task Manager API")
}
func healthHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintln(w, "Server is healthy")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	config = loadConfig()
	if config.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET is not set in .env file")
	}
	if config.DBHost == "" ||
    config.DBPort == "" ||
    config.DBUser == "" ||
    config.DBPassword == "" ||
    config.DBName == "" {
    log.Fatal("Database configuration is incomplete")
	}
	if err:= connectDB(); err!= nil{
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database successfully")	
	taskrepo := repository.TaskRepository{DB:DB}
	taskservice := service.Taskservice{TaskRepo:taskrepo}
	userrepo := repository.UserRepo{DB:DB}
	userservice := service.UserService{UserRepo:userrepo}
	handler := Handler{TaskService:taskservice, TaskRepo:taskrepo, UserService:userservice}
	r:= gin.Default()
	r.GET("/tasks",AuthMiddleware,handler.handleGetTask)
	r.GET("/tasks/:id",AuthMiddleware,handler.handleGetTask)
	// protected := r.Group("/tasks")
	// protected.Use(AuthMiddleware)
	// protected.GET("",handleGetTask)
	r.POST("/tasks",AuthMiddleware,handler.handlePostTask)
	r.POST("/register",handler.handleRegister)
	r.POST("/login",handler.handleLogin)
	r.PUT("/tasks/:id",AuthMiddleware,handler.handlePutTask)
	r.DELETE("/tasks/:id",AuthMiddleware,handler.handleDeleteTask)
	fmt.Println("Server is running on port 8080")
	r.Run(":"+config.PORT)
}