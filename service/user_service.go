package service
import(
	"errors"
	"golang.org/x/crypto/bcrypt"
	"task-manager-api/repository"
	"strings"
	"task-manager-api/models"
)
type UserService struct{
	UserRepo repository.UserRepo
}
var ErrEmptyEmail = errors.New("Email cannot be empty")
var ErrEmptyPassword = errors.New("Password cannot be empty")
var PasswordTooShort = errors.New("Password must be at least 6 characters")
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
		return models.User{},err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(password))
	if err != nil {
		return models.User{},err
	}
	return user,nil
}