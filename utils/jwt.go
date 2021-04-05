package utils

import (
  "errors"
  "log"
  "net/http"

  "github.com/dgrijalva/jwt-go"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "golang.org/x/crypto/bcrypt"
)

// Claims Token interface
type Claims struct {
  ID        primitive.ObjectID `json:"_id" bson:"_id"`
  Username  string             `json:"username" bson:"username"`
  Role      int                `json:"role_key" bson:"role_key"`
  UserAgent string             `json:"user_agent" bson:"user_agent`
  jwt.StandardClaims
}

var JwtKey = []byte("secret")
var Header = "Authorization"

// ComparePassword , compare plain password with the hashed password
func ComparePassword(plainPassword []byte, hashedPassword string) bool {

  hashedPass := []byte(hashedPassword)
  result := bcrypt.CompareHashAndPassword(hashedPass, plainPassword)

  if result != nil {
    log.Println(result.Error())
    return false
  }

  return true

}

// HashSaltPassword , generate salt and hashing password
func HashSaltPassword(pwd []byte) string {
  hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)

  if err != nil {
    log.Println(err)
  }
  return string(hash)
}

// CheckAuthToken func
func CheckAuthToken(token string, role []int, r *http.Request) error {
  claims := &Claims{}
  _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
    return JwtKey, nil
  })
  if !ArrContains(role, claims.Role) || r.Header.Get("User-Agent") != claims.UserAgent {
    return errors.New("you dont have permission to access")
  }
  if err != nil {
    return errors.New("invalid user token")
  }
  return nil
}

func ArrContains(slice []int, value int) bool {
  for _, s := range slice {
    if s == value {
      return true
    }
  }
  return false
}
