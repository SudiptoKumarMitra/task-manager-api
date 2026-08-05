package main
import(
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)
var DB *sql.DB
func connectDB()error {
	connectionString:="host=localhost port=5432 user=postgres password=sudipto dbname=taskmanager sslmode=disable"
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