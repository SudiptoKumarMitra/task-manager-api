package main

import(
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(c *gin.Context){
	authorization := c.GetHeader("Authorization")
	if !strings.HasPrefix(authorization, "Bearer ") {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message" : "Unauthorized",
 		})
		c.Abort()
		return
	}
	tokenString := strings.TrimPrefix(authorization,"Bearer ")
	claims,err := validateToken(tokenString)
		if err != nil {
		c.JSON(http.StatusUnauthorized,gin.H{
			"message" : "Unauthorized",
 		})
		c.Abort()
		return
	}
	userIDvalue,ok := claims["user_id"].(float64)
	if !ok{
		c.JSON(http.StatusUnauthorized,gin.H{
			"message" : "Unauthorized",
 		})
		c.Abort()
		return
	}
	userID := int(userIDvalue)
	c.Set("user_id",userID)
	c.Next()
} 