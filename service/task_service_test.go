package service

import (
	"errors"
	"task-manager-api/apperrors"
	"task-manager-api/models"
	"testing"
)

type FakeTaskRepository struct {
	CreateTaskCalled bool
	CreatedTitle     string
	CreateduserID    int

	GetTasksByIDCalled bool
	GetTasksByIDID     int
	GetTasksByIDuserID int

	GetTaskByIDError error

	UpdateTaskCalled bool
	UpdateTaskTitle  string
	UpdateTaskID     int
	UpdateTaskUserID int

	UpdateTaskResult bool
	UpdateTaskError  error

	DeleteTaskCalled bool
	DeleteTaskID     int
	DeleteTaskUserID int

	DeleteTaskResult bool
	DeleteTaskError  error
}

func (f *FakeTaskRepository) CreateTask(title string, userID int) (models.Task, error) {
	f.CreateTaskCalled = true
	f.CreatedTitle = title
	f.CreateduserID = userID
	return models.Task{
		Title: title,
		ID:    1,
	}, nil
}
func (f *FakeTaskRepository) GetTasksByUser(userID int) ([]models.Task, error) {
	return nil, nil
}

func (f *FakeTaskRepository) GetTaskByID(id int, userID int) (models.Task, error) {
	f.GetTasksByIDCalled = true
	f.GetTasksByIDID = id
	f.GetTasksByIDuserID = userID
	if f.GetTaskByIDError != nil {
		return models.Task{}, f.GetTaskByIDError
	}
	return models.Task{
		Title: "Learning Testing",
		ID:    id,
	}, nil
}

func (f *FakeTaskRepository) UpdateTask(title string, id int, userID int) (bool, error) {
	f.UpdateTaskCalled = true
	f.UpdateTaskTitle = title
	f.UpdateTaskID = id
	f.UpdateTaskUserID = userID
	if f.UpdateTaskError != nil {
		return false, f.UpdateTaskError
	}
	return f.UpdateTaskResult, nil
}

func (f *FakeTaskRepository) DeleteTask(id int, userID int) (bool, error) {
	f.DeleteTaskCalled = true
	f.DeleteTaskID = id
	f.DeleteTaskUserID = userID
	return f.DeleteTaskResult, f.DeleteTaskError
}
func TestCreateTask_EmptyTitle(t *testing.T) {
	fakeRepo := &FakeTaskRepository{}
	taskservice := Taskservice{TaskRepo: fakeRepo}
	_, err := taskservice.CreateTask("   ", 1)

	if !errors.Is(err, apperrors.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
	if fakeRepo.CreateTaskCalled {
		t.Errorf("expected CreateTaskCalled to be false")
	}
}
func TestCreateTask_Success(t *testing.T) {
	fakerepo := &FakeTaskRepository{}
	service := Taskservice{TaskRepo: fakerepo}
	task, err := service.CreateTask("Jion", 2)

	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if task.ID != 1 {
		t.Errorf("expected ID 1, got %d", task.ID)
	}

	if task.Title != "Jion" {
		t.Errorf("expected title Learning Testing, got %s", task.Title)
	}
	if fakerepo.CreatedTitle != task.Title {
		t.Errorf("expected title same but got %s", fakerepo.CreatedTitle)
	}
	if fakerepo.CreateduserID != 2 {
		t.Errorf("expected userID same but got %d", fakerepo.CreateduserID)
	}
	if !fakerepo.CreateTaskCalled {
		t.Errorf("expected CreateTaskCalled to be true")
	}
}
func TestGetTaskByID_Success(t *testing.T) {
	fakerepo := &FakeTaskRepository{}
	service := Taskservice{TaskRepo: fakerepo}
	task, err := service.GetTaskByID(8, 9)
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if fakerepo.GetTasksByIDID != 8 {
		t.Errorf("expected GetTasksByID 8, got %d", fakerepo.GetTasksByIDID)
	}
	if fakerepo.GetTasksByIDuserID != 9 {
		t.Errorf("expected GetTasksByuserID 9, got %d", fakerepo.GetTasksByIDuserID)
	}
	if task.ID != 8 {
		t.Errorf("expected ID 8, got %d", task.ID)
	}
	if task.Title != "Learning Testing" {
		t.Errorf("expected Title Learning Testing, got %s", task.Title)
	}
	if !fakerepo.GetTasksByIDCalled {
		t.Errorf("expected GetTasksByUserCalled to be true")
	}
}

func TestGetTaskBtID_Error(t *testing.T) {
	fakerepo := &FakeTaskRepository{}
	fakerepo.GetTaskByIDError = apperrors.ErrTaskNotFound
	service := Taskservice{TaskRepo: fakerepo}
	_, err := service.GetTaskByID(8, 9)
	if !errors.Is(err, apperrors.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestUpdateTask_Success(t *testing.T) {
	fakerepo := &FakeTaskRepository{}
	fakerepo.UpdateTaskResult = true
	service := Taskservice{TaskRepo: fakerepo}
	check, err := service.UpdateTask("Learning Testing", 8, 9)
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if fakerepo.UpdateTaskTitle != "Learning Testing" {
		t.Errorf("expected title Learning Testing, got %s", fakerepo.UpdateTaskTitle)
	}
	if fakerepo.UpdateTaskID != 8 {
		t.Errorf("expected id 8, got %d", fakerepo.UpdateTaskID)
	}
	if fakerepo.UpdateTaskUserID != 9 {
		t.Errorf("expected userID 9, got %d", fakerepo.UpdateTaskUserID)
	}
	if !check {
		t.Errorf("expected check to be true")
	}
	if !fakerepo.UpdateTaskCalled {
		t.Errorf("expected UpdateTaskCalled to be true")
	}
}
func TestUpdateTask_EmptyTitle(t *testing.T) {
	fakeRepo := &FakeTaskRepository{}
	taskservice := Taskservice{TaskRepo: fakeRepo}
	_, err := taskservice.UpdateTask("   ", 1, 2)

	if !errors.Is(err, apperrors.ErrEmptyTitle) {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
	if fakeRepo.UpdateTaskCalled {
		t.Errorf("expected UpdateTaskCalled to be false")
	}
}
func TestUpdateTask_NotFound(t *testing.T) {
	fakeRepo := &FakeTaskRepository{}
	fakeRepo.UpdateTaskResult = false
	serviceservice := Taskservice{TaskRepo: fakeRepo}
	check, err := serviceservice.UpdateTask("Learning Testing", 8, 9)
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if check {
		t.Errorf("expected check to be false")
	}
}
func TestUpdateTask_Error(t *testing.T) {
	fakeRepo := &FakeTaskRepository{}
	fakeRepo.UpdateTaskError = errors.New("database error")
	service := Taskservice{TaskRepo: fakeRepo}
	check, err := service.UpdateTask("Learning Testing", 8, 9)
	if err != fakeRepo.UpdateTaskError {
		t.Errorf("expected databse error got %v", err)
	}
	if check {
		t.Errorf("expected check to be false")
	}
}
func TestDeleteTask_Success(t *testing.T) {
	fakerepo := &FakeTaskRepository{}
	fakerepo.DeleteTaskResult = true
	service := Taskservice{TaskRepo: fakerepo}
	check, err := service.DeleteTask(1, 2)
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if fakerepo.DeleteTaskID != 1 {
		t.Errorf("expected id 1, got %d", fakerepo.DeleteTaskID)
	}
	if fakerepo.DeleteTaskUserID != 2 {
		t.Errorf("expected userID 2, got %d", fakerepo.DeleteTaskUserID)
	}
	if !check {
		t.Errorf("expected check to be true")
	}
	if !fakerepo.DeleteTaskCalled {
		t.Errorf("expected DeleteTaskCalled to be true")
	}
}
func TestDeleteTask_NotFound(t *testing.T) {
	fakerepo := &FakeTaskRepository{}
	fakerepo.DeleteTaskResult = false
	service := Taskservice{TaskRepo: fakerepo}
	check, err := service.DeleteTask(1, 2)
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if check {
		t.Errorf("expected check to be false")
	}
}
func TestDeleteTaskError(t *testing.T) {
	fakerepo := new(FakeTaskRepository)
	fakerepo.DeleteTaskError = errors.New("database error")
	service := Taskservice{TaskRepo: fakerepo}
	check, err := service.DeleteTask(1, 2)
	if !errors.Is(err, fakerepo.DeleteTaskError) {
		t.Errorf("expected error got %v", err)
	}
	if check {
		t.Errorf("expected check to be false")
	}
}
