package main
import (
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"
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