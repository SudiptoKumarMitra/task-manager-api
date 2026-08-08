package main
import (
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"
	"golang.org/x/crypto/bcrypt"
)

func handleGetTask(c *gin.Context) {
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
		var task Task
		err = DB.QueryRow("SELECT id,title FROM tasks WHERE id = $1 AND user_id = $2",id,userID).Scan(&task.ID,&task.Title)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Task not found",
			})
			return
		}
		c.JSON(http.StatusOK,task)
	} else {
		rows,err := DB.Query("SELECT id,title FROM tasks WHERE user_id = $1;",userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to connect with database",
			})
			return
		}
		defer rows.Close()
		var tasks []Task
		for rows.Next() {
			var task Task
			err := rows.Scan(&task.ID, &task.Title)
				if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error in fetching data",
			})
				return
			}
			tasks=append(tasks,task)
		}
		c.JSON(http.StatusOK,tasks)
	}
}

func handlePostTask(c *gin.Context) {
	var task Task
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
	var id int
	err = DB.QueryRow("INSERT INTO tasks (title, user_id) VALUES ($1,$2) RETURNING id",task.Title,userID).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Failed to Insert into database",
		})
		return
	}
	task.ID = id
	c.JSON(http.StatusCreated,task)
}

func handlePutTask(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "Invalid ID",
		})
		return
	}
	var task Task
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
	res,err := DB.Exec("UPDATE tasks SET title = $1 WHERE id = $2 AND user_id = $3",task.Title,id,userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Error in updating data",
		})
		return
	}
	affected,err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Error in database",
		})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound,gin.H{
			"error" : "NO task exists with this ID",
		})
		return
	}
	task.ID=id
	c.JSON(http.StatusOK,task)
}

func handleDeleteTask(c *gin.Context) {
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
	res,err := DB.Exec("DELETE FROM tasks WHERE id = $1 AND  user_id = $2",id,userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Error Occured in database",
		})
		return
	}
	affected,err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "Error Occured in database",
		})
		return
	}
	if affected == 0 {
		c.JSON(http.StatusNotFound,gin.H{
			"error" : "Task not found",
		})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"message" : "Task Deleted",
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
	tokenString,err := generateToken(userID,role)
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