package main
import (
	"github.com/golang-jwt/jwt/v5"
	"fmt"
)

func generateToken(user_id int, role string) (string , error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"user_id": user_id,
		"role" : role,
	})
	signed_token,err:= token.SignedString([]byte("my-secret"))
	if err != nil {
		return "",err
	}
	return signed_token,nil
}
func validateToken(tokenString string) (jwt.MapClaims , error){
	token,err := jwt.Parse(tokenString, func (token *jwt.Token) (interface{}, error){
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("invalid signing algorithm") 
			
		}
		return ([]byte("my-secret")) , nil
	})
	if err != nil{
		return nil, fmt.Errorf("Invalid Token")
	}
	if !token.Valid {
		return nil, fmt.Errorf("Invalid Token")
		
	}
	claims,ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Invalid Claims")
	}
	return claims,nil
}