package main
import (
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"
	"golang.org/x/crypto/bcrypt"
	"task-manager-api/repository"
	"task-manager-api/models"
	"task-manager-api/service"
	"errors"
)
 type Handler struct {
	TaskService service.Taskservice
	TaskRepo repository.TaskRepository	
}
func (h *Handler) handleGetTask(c *gin.Context) {
	userIDValue,ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID,ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	idstr:=c.Param("id")
	if idstr != "" {
		id,err:=strconv.Atoi(idstr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID",
			})
			return
		}
		task,err := h.TaskService.GetTaskByID(id,userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
			})
			return
		}
		c.JSON(http.StatusOK,task)
	} else {
		tasks,err := h.TaskService.GetTasksByUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to connect with database",
			})
			return
		}
		
		c.JSON(http.StatusOK,tasks)
	}
}

func (h *Handler) handlePostTask(c *gin.Context) {
	var task models.Task
	err := c.ShouldBindJSON(&task)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid JSON",
		})
		return
	}
	userIDValue,ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID,ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	createdTask,err := h.TaskService.CreateTask(task.Title,userID)
	if err != nil {
		if errors.Is(err, service.ErrEmptyTitle) {
			c.JSON(http.StatusBadRequest,gin.H{
				"error" : "Title cannot be empty",
			})
			return
		}
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Failed to Insert into database",
		})
		return
	}
	c.JSON(http.StatusCreated,createdTask)
}

func (h *Handler) handlePutTask(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid ID",
		})
		return
	}
	var task models.Task
	err = c.ShouldBindJSON(&task)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid JSON",
		})
		return
	}
	userIDValue,ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID,ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	check,err := h.TaskService.UpdateTask(task.Title,id,userID)
	if err != nil {
		if errors.Is(err,service.ErrEmptyTitle) {
			c.JSON(http.StatusBadRequest,gin.H{
				"error" : "Title cannot be empty",
			})
			return
		}
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Failed to update task",
    })
    return
	}

	if !check {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}
	task.ID = id	
	c.JSON(http.StatusOK,task)
}

func (h *Handler) handleDeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid ID",
		})
		return
	}
	userIDValue,ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID,ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	deleted,err := h.TaskService.DeleteTask(id,userID)
	if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Error occurred in database",
    })
    return
	}

	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task Deleted",
	})
}
func handleRegister(c *gin.Context) {
	var request RegisterRequest
	err := c.ShouldBindJSON(&request)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid JSON",
		})
		return
	}
	hasedPassword,err := bcrypt.GenerateFromPassword([]byte(request.Password),bcrypt.DefaultCost)
	if err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Failed to generate password",
		})
		return
	}
	_, err = DB.Exec("INSERT INTO users (email,password) VALUES ($1,$2)",request.Email,string(hasedPassword))
	if err != nil{
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Failed to insert into database",
		})
		return
	}
	c.JSON(http.StatusCreated,gin.H{
		"message" : "User Created",
	})
}
func handleLogin(c *gin.Context) {
	var req LoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid JSON",
		})
		return
	}
	var hashedPassword,role string
	var userID int
	err = DB.QueryRow("SELECT id, password, role FROM users WHERE email = $1",req.Email).Scan(&userID,&hashedPassword,&role)
	if err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"error" : "Invalid Credentials",
		})
		return
	}
	err =bcrypt.CompareHashAndPassword([]byte(hashedPassword),[]byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"error" : "Invalid Credentials",
		})
		return
	}
	tokenString,err := generateToken(userID,role,config.JWT_SECRET)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Failed to generate token",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"token" : tokenString,
	})
}