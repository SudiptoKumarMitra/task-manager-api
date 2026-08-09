package repository
import (
	"database/sql"
	"task-manager-api/models"
)
type TaskRepository struct {
	DB *sql.DB
}

func (r *TaskRepository) GetTasksByUser (userID int) ([]models.Task,error){
	var tasks []models.Task
	rows,err := r.DB.Query("SELECT id, title FROM tasks WHERE user_id =$1",userID)
	if err != nil {
		return nil,err
	}
	defer rows.Close()
	for rows.Next(){
		var task models.Task
		err := rows.Scan(&task.ID,&task.Title)
		if err != nil {
			return nil,err
		}
		
		tasks = append(tasks,task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks,nil
}
func (r * TaskRepository) GetTaskByID(id int, userID int) (models.Task,error){
	var task models.Task
	err := r.DB.QueryRow("SELECT id, title FROM tasks WHERE id = $1 AND user_id = $2",id,userID).Scan(&task.ID,&task.Title)
	if err != nil {
		return task,err
	}
	return task,nil
}

func (r * TaskRepository) CreateTask(title string, userID int) (models.Task,error){
	var task models.Task
	task.Title = title
	err := r.DB.QueryRow("INSERT INTO tasks (title, user_id) VALUES ($1,$2)  RETURNING id",title,userID).Scan(&task.ID)
	if err != nil {
		return task,err
	}
	return task,nil
}
func (r * TaskRepository) UpdateTask(title string, id int, userID int) (bool,error){
	res,err := r.DB.Exec("UPDATE tasks SET title = $1 WHERE id = $2 AND user_id = $3",title,id,userID)
	if err != nil {
		return false,err
	}
	rows,err := res.RowsAffected()
	if err != nil {
		return false,err
	}
	return rows == 1,nil
}
func (r * TaskRepository) DeleteTask(id int, userID int) (bool,error){
	res,err := r.DB.Exec("DELETE FROM tasks WHERE id = $1 AND  user_id = $2",id,userID)
	if err != nil {
		return false,err
	}
	affected,err := res.RowsAffected()
	if err != nil {
		return false,err
	}
	return affected == 1,nil
}