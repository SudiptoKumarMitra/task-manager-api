package service
import (
	"task-manager-api/repository"
	"task-manager-api/models"
	"strings"
)
type Taskservice struct{
	TaskRepo repository.TaskRepository
}
func (S *Taskservice) GetTasksByUser(userID int) ([]models.Task,error){
	return S.TaskRepo.GetTasksByUser(userID)
}
func (S *Taskservice) GetTaskByID(id int, userID int) (models.Task,error){
	return S.TaskRepo.GetTaskByID(id,userID)
}
func (S *Taskservice) CreateTask(title string, userID int) (models.Task,error){
	title = strings.TrimSpace(title)
	if title == "" {
		return models.Task{},ErrEmptyTitle
	}
	return S.TaskRepo.CreateTask(title,userID)
}

func (S *Taskservice) UpdateTask(title string, id int, userID int) (bool,error){
	title = strings.TrimSpace(title)
	if title == "" {
		return false,ErrEmptyTitle
	}
	return S.TaskRepo.UpdateTask(title,id,userID)
}
func (S *Taskservice) DeleteTask(id int, userID int) (bool,error){
	return S.TaskRepo.DeleteTask(id,userID)
}