package main
import(
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"fmt"
)
var DB *sql.DB
func connectDB()error {
	connectionString := fmt.Sprintf(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    config.DBHost,
    config.DBPort,
    config.DBUser,
    config.DBPassword,
    config.DBName,
)
	var err error
	DB,err=sql.Open("pgx",connectionString)
	if err != nil {
		return err
	}
	err=DB.Ping()
	if err != nil {
		return err
	}
	return nil
	
}