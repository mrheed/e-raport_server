package routehandler

import (
  "context"
  "encoding/json"
  "net/http"
  "time"

  "github.com/dgrijalva/jwt-go"
  db "github.com/syahidnurrohim/restapi/database"
  mod "github.com/syahidnurrohim/restapi/models"
  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

/* ------------------
Permission details
1. Administrator
2. Guru PKN
3. Guru Agama
4. Guru Biasa
5. Wali Kelas
-------------------- */
var AllUser = []int{1, 2, 3, 4, 5}
var Admin = []int{1}
var GuruPKN = []int{2}
var GuruAgama = []int{3}
var GuruBiasa = []int{4}
var Wali = []int{5}

// SignInCridentials interface
type SignInCridentials struct {
  ID       primitive.ObjectID `json:"_id" bson:"_id"`
  Username string             `json:"username" bson:"username"`
  Password string             `json:"password" bson:"password"`
  RoleKey  int                `json:"role_key,omitempty" bson:"role_key,omitempty"`
}

// SignInRequestHandler method
func SignInRequestHandler(w http.ResponseWriter, r *http.Request) {

  var cridentials SignInCridentials
  var userCridentials UserData
  var plainPassword string

  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  err := json.NewDecoder(r.Body).Decode(&cridentials)
  if err != nil {
    throw.Error(err.Error())
    return
  }
  plainPassword = cridentials.Password
  if err := db.User.FindOne(context.Background(), bson.M{"username": cridentials.Username}).Decode(&cridentials); err != nil {
    throw.Error(err.Error())
    return
  }
  if !tool.ComparePassword([]byte(plainPassword), cridentials.Password) {
    throw.Error("user tidak ditemukan")
    return
  }
  if err := db.User.FindOne(context.Background(), bson.M{
    "username": cridentials.Username,
    "password": cridentials.Password,
    "_id":      cridentials.ID,
  }).Decode(&cridentials); err != nil {
    throw.Error(err.Error())
  }

  expirationTime := time.Now().Add(5 * time.Hour)

  claims := &tool.Claims{
    ID:        cridentials.ID,
    Username:  cridentials.Username,
    Role:      cridentials.RoleKey,
    UserAgent: r.Header.Get("User-Agent"),
    StandardClaims: jwt.StandardClaims{
      ExpiresAt: expirationTime.Unix(),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString(tool.JwtKey)

  if err != nil {
    throw.Error(err.Error())
    return
  }

  db.Session.InsertOne(context.Background(), bson.M{
    "_id":          primitive.NewObjectID(),
    "user_id":      cridentials.ID,
    "username":     cridentials.Username,
    "role_key":     cridentials.RoleKey,
    "token":        tokenString,
    "date_added":   time.Now(),
    "date_expired": expirationTime,
  })

  db.User.UpdateOne(context.Background(), bson.M{"_id": cridentials.ID}, bson.M{
    "$set": bson.M{"last_login": time.Now()},
  })
  db.User.FindOne(context.Background(), bson.M{"_id": cridentials.ID}).Decode(&userCridentials)

  payload := map[string]interface{}{
    "user_data": &UserData{
      ID:       cridentials.ID,
      Username: cridentials.Username,
      Name: UserFName{
        Firstname: userCridentials.Name.Firstname,
        Lastname:  userCridentials.Name.Lastname,
      },
      RoleKey:   userCridentials.RoleKey,
      LastLogin: userCridentials.LastLogin,
    },
    "http_data": &http.Cookie{
      Name:    "token",
      Value:   tokenString,
      Expires: expirationTime,
    },
  }
  throw.Response(payload)
}

// PurgeSingleToken , remove single session from database
func PurgeSingleToken(w http.ResponseWriter, r *http.Request) {
  if splittedToken, err := tool.VerifyHeader(AllUser, r, w); err == nil {
    db.Session.FindOneAndDelete(context.Background(), bson.M{"token": splittedToken})
  }
}
