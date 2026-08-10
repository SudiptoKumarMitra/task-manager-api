package repository
import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgconn"
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