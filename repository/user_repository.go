package repository
import (
	"database/sql"
	"task-manager-api/models"
	"github.com/jackc/pgx/v5/pgconn"
	"errors"
)
type UserRepo struct{
 DB *sql.DB
}
var ExistEmailError = errors.New("email already exists")
func (r *UserRepo) RegisterUser(email string, hashedPassword string) error{
	_,err := r.DB.Exec("INSERT INTO users (email,password) VALUES ($1,$2)",email,hashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ExistEmailError
			}
		}
		return err
	}
	return nil
}
func (r *UserRepo) GetUserByEmail(email string) (models.User,error){
	var user models.User
	err := r.DB.QueryRow("SELECT id, email, password, role From users WHERE email = $1",email).Scan(&user.ID,&user.Email,&user.Password,&user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{},err
		}
		return user,err
	}
	return user,nil
}