package service
import(
	"golang.org/x/crypto/bcrypt"
	"task-manager-api/repository"
	"strings"
	"task-manager-api/models"
	"database/sql"
)
type UserService struct{
	UserRepo repository.UserRepo
}

func (S *UserService) RegisterUser(email string, password string) error{
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmptyEmail
	}
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 6 {
		return PasswordTooShort
	}
	hasedPassword,err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return S.UserRepo.RegisterUser(email,string(hasedPassword))
}
func (S *UserService) LoginUser(email string, password string) (models.User,error){
	email = strings.TrimSpace(email)
	if email == "" {
		return models.User{},ErrEmptyEmail
	}
	if password == "" {
		return models.User{},ErrEmptyPassword
	}
	if len(password) < 6 {
		return models.User{},PasswordTooShort
	}
	user,err := S.UserRepo.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{},ErrInvalidCredentials
		}
		return models.User{},err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(password))
	if err != nil {
		return models.User{},ErrInvalidCredentials
	}
	return user,nil
}