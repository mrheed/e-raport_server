package routehandler

import (
  "context"
  "log"
  "net/http"
  "time"

  "github.com/dgrijalva/jwt-go"
  db "github.com/syahidnurrohim/restapi/database"
  mod "github.com/syahidnurrohim/restapi/models"
  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

// RefreshPayload Interface
type RefreshPayload struct {
  Token         string
  AccessGranted bool
  ExpiresAt     time.Time
  IsError       bool
  ErrorMessage  string
}

// ErrorRefresh Interface
type ErrorRefresh struct {
  IsError         bool
  ErrorResponse   string
  ErrorStatusCode int
}

// SessionStruct interface
type SessionStruct struct {
  ID          primitive.ObjectID `json:"_id" bson:"_id"`
  UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
  Token       string             `json:"token" bson:"token"`
  Role        int                `json:"role_key" bson:"role_key"`
  UserAgent   string             `json:"user_agent" bson:"user_agent"`
  Username    string             `json:"username" bson:"username"`
  DateAdded   time.Time          `json:"date_added" bson:"date_added"`
  DateExpired time.Time          `json:"date_expired" bson:"date_expired"`
}

func clearSession(id primitive.ObjectID, token string) {
  db.Session.DeleteMany(context.Background(), bson.M{"user_id": id, "token": token})
}

// Refresh function to reload token
func Refresh(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  splittedToken, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w)
  if err != nil {
    throw.Error(err.Error())
    return
  }
  var DeletedSession SessionStruct
  claims := &tool.Claims{}
  tkn, parseErr := jwt.ParseWithClaims(splittedToken, claims, func(token *jwt.Token) (interface{}, error) {
    return tool.JwtKey, nil
  })
  if parseErr != nil {
    clearSession(claims.ID, splittedToken)
    throw.Error(parseErr.Error())
    return
  }
  if r.Header.Get("User-Agent") != claims.UserAgent {
    clearSession(claims.ID, splittedToken)
    throw.Error("invalid token provided")
    return
  }

  expirationTime := time.Now().Add(5 * time.Hour)
  claims.ExpiresAt = expirationTime.Unix()

  if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 30*time.Second {
    clearSession(claims.ID, splittedToken)
    throw.Error("token already expired")
    return
  }
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString(tool.JwtKey)
  if err != nil {
    throw.Error(err.Error())
  }
  db.Session.FindOneAndDelete(context.Background(), bson.M{
    "user_id": claims.ID,
    "token":   splittedToken,
  }).Decode(&DeletedSession)
  if DeletedSession.UserID != claims.ID || !tkn.Valid {
    throw.Error("invalid token provided")
    return
  }
  db.Session.InsertOne(context.Background(), bson.M{
    "_id":          primitive.NewObjectID(),
    "user_agent":   claims.UserAgent,
    "username":     claims.Username,
    "date_expired": expirationTime,
    "token":        tokenString,
    "role_key":     claims.Role,
    "date_added":   time.Now(),
    "user_id":      claims.ID,
  })
  throw.Response(bson.M{
    "token":      tokenString,
    "expires_at": expirationTime,
  })
}

// CheckExpireTkn , checking user sessin in db
func CheckExpireTkn() {

  for range time.Tick(time.Second * 1) {
    var sessions []string
    var users []string
    result, err := db.Session.Find(context.Background(), bson.D{})
    if err != nil {
      return
    }
    for result.Next(context.Background()) {
      var session SessionStruct
      result.Decode(&session)
      sessions = append(sessions, session.UserID.String())
      users = append(users, session.Username)
      if session.DateExpired.Local().Unix() < time.Now().Local().Unix() {
        db.Session.DeleteOne(context.Background(), bson.M{"_id": session.ID})
        log.Println(session.Username, "has logged out")
      }
    }
    if len(sessions) == 0 {
      log.Println("No user online")
    }
    sessions = sessions[:0]

    result.Close(context.Background())

  }

  time.Sleep(time.Millisecond * 200)

}
