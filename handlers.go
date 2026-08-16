package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"task-manager-api/apperrors"
	"task-manager-api/models"
	"task-manager-api/service"
	"task-manager-api/utils"
)

type Handler struct {
	TaskService service.Taskservice
	UserService service.UserService
}

func (h *Handler) handleGetTask(c *gin.Context) {
	userIDValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	idstr := c.Param("id")
	if idstr != "" {
		id, err := strconv.Atoi(idstr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid ID",
			})
			return
		}
		task, err := h.TaskService.GetTaskByID(id, userID)
		if err != nil {
			if errors.Is(err, apperrors.ErrTaskNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Task not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to connect with database",
			})
			return
		}
		c.JSON(http.StatusOK, task)
	} else {
		tasks, err := h.TaskService.GetTasksByUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to connect with database",
			})
			return
		}

		c.JSON(http.StatusOK, tasks)
	}
}

func (h *Handler) handlePostTask(c *gin.Context) {
	var task models.Task
	err := c.ShouldBindJSON(&task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON",
		})
		return
	}
	userIDValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	createdTask, err := h.TaskService.CreateTask(task.Title, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmptyTitle) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Title cannot be empty",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to Insert into database",
		})
		return
	}
	c.JSON(http.StatusCreated, createdTask)
}

func (h *Handler) handlePutTask(c *gin.Context) {
	idstr := c.Param("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}
	var task models.Task
	err = c.ShouldBindJSON(&task)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON",
		})
		return
	}
	userIDValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	check, err := h.TaskService.UpdateTask(task.Title, id, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmptyTitle) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Title cannot be empty",
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
	c.JSON(http.StatusOK, task)
}

func (h *Handler) handleDeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}
	userIDValue, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	deleted, err := h.TaskService.DeleteTask(id, userID)
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
func (h *Handler) handleRegister(c *gin.Context) {
	var request models.RegisterRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		validationmessage := utils.FormatValidationErrors(err)
		if len(validationmessage) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": validationmessage,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON",
		})
		return
	}
	err = h.UserService.RegisterUser(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrEmptyEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email cannot be empty",
			})
			return
		}
		if errors.Is(err, apperrors.ErrEmptyPassword) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password cannot be empty",
			})
			return
		}
		if errors.Is(err, apperrors.PasswordTooShort) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must be at least 6 characters",
			})
			return
		}
		if errors.Is(err, apperrors.ExistEmailError) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to insert into database",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User Created",
	})
}
func (h *Handler) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		validationmessage := utils.FormatValidationErrors(err)
		if len(validationmessage) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": validationmessage,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON",
		})
		return
	}
	user, err := h.UserService.LoginUser(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Credentials",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
		return
	}
	tokenString, err := generateToken(user.ID, user.Role, config.JWT_SECRET)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
	})
}
