package repository
import (
	"database/sql"
	"task-manager-api/models"
)
type UserRepo struct{
 DB *sql.DB
}
func (r *UserRepo) RegisterUser(email string, hashedPassword string) error{
	_,err := r.DB.Exec("INSERT INTO users (email,password) VALUES ($1,$2)",email,hashedPassword)
	if err != nil {
		return err
	}
	return nil
}
func (r *UserRepo) GetUserByEmail(email string) (models.User,error){
	var user models.User
	err := r.DB.QueryRow("SELECT id, email, password, role From users WHERE email = $1",email).Scan(&user.ID,&user.Email,&user.Password,&user.Role)
	if err != nil {
		return user,err
	}
	return user,nil
}