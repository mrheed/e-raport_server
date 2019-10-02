package routehandler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	db "github.com/syahidnurrohim/restapi/database"
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
	Username    string             `json:"username" bson:"username"`
	DateAdded   time.Time          `json:"date_added" bson:"date_added"`
	DateExpired time.Time          `json:"date_expired" bson:"date_expired"`
}

// Refresh function to reload token
func Refresh(w http.ResponseWriter, r *http.Request) {
	if splittedToken, err := tool.VerifyHeader(AllUser, r, w); err == nil {
		var Error ErrorRefresh
		var Payload RefreshPayload
		var DeletedSession SessionStruct

		claims := &tool.Claims{}
		tkn, parseErr := jwt.ParseWithClaims(splittedToken, claims, func(token *jwt.Token) (interface{}, error) {
			return tool.JwtKey, nil
		})
		if r.Header.Get("User-Agent") != claims.UserAgent {
			return
		}
		expirationTime := time.Now().Add(5 * time.Hour)
		claims.ExpiresAt = expirationTime.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(tool.JwtKey)

		db.Session.FindOneAndDelete(context.Background(), bson.D{
			primitive.E{Key: "user_id", Value: claims.ID},
			primitive.E{Key: "token", Value: splittedToken},
		}).Decode(&DeletedSession)

		if DeletedSession.UserID == claims.ID {

			db.Session.InsertOne(context.Background(),
				bson.D{
					primitive.E{Key: "_id", Value: primitive.NewObjectID()},
					primitive.E{Key: "user_id", Value: claims.ID},
					primitive.E{Key: "username", Value: claims.Username},
					primitive.E{Key: "date_added", Value: time.Now()},
					primitive.E{Key: "date_expired", Value: expirationTime},
					primitive.E{Key: "token", Value: tokenString},
				})

		} else {
			Error = ErrorRefresh{
				IsError:         true,
				ErrorResponse:   `{"message" : "Invalid Token Provided"}`,
				ErrorStatusCode: http.StatusMovedPermanently,
			}
		}

		if !tkn.Valid {
			Error = ErrorRefresh{
				IsError:         true,
				ErrorResponse:   `{"message" : "Invalid Token Provided"}`,
				ErrorStatusCode: http.StatusMovedPermanently,
			}
		}

		if (parseErr != nil) || (err != nil) {
			if parseErr == jwt.ErrSignatureInvalid {
				Error = ErrorRefresh{
					IsError:         true,
					ErrorResponse:   `{"message" : "Invalid Token Signature"}`,
					ErrorStatusCode: http.StatusMovedPermanently,
				}
			}
			Error = ErrorRefresh{
				IsError:         true,
				ErrorResponse:   `{"message" : "Opps something went wrong"}`,
				ErrorStatusCode: http.StatusBadRequest,
			}
		}

		if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 30*time.Second {
			Error = ErrorRefresh{
				IsError:         true,
				ErrorResponse:   `{"message" : "Token already expired"}`,
				ErrorStatusCode: http.StatusMovedPermanently,
			}
		}

		if Error.IsError {

			w.WriteHeader(Error.ErrorStatusCode)
			Payload = RefreshPayload{
				AccessGranted: false,
				IsError:       true,
				ErrorMessage:  Error.ErrorResponse,
			}

			json.NewEncoder(w).Encode(Payload)

			return
		}

		Payload = RefreshPayload{
			Token:         tokenString,
			AccessGranted: true,
			ExpiresAt:     expirationTime,
		}

		json.NewEncoder(w).Encode(Payload)
	}
}

// CheckExpireTkn , checking user sessin in db
func CheckExpireTkn() {

	for range time.Tick(time.Second * 1) {

		var sessions []string
		var users []string
		mutex := &sync.Mutex{}
		mutex.Lock()
		result, err := db.Session.Find(context.Background(), bson.D{})
		mutex.Unlock()
		if err != nil {
			return
		}

		for result.Next(context.Background()) {
			var session SessionStruct
			mutex.Lock()
			result.Decode(&session)
			mutex.Unlock()
			mutex.Lock()
			sessions = append(sessions, session.UserID.String())
			mutex.Unlock()
			mutex.Lock()
			users = append(users, session.Username)
			mutex.Unlock()
			if session.DateExpired.Local().Unix() < time.Now().Local().Unix() {
				mutex.Lock()
				db.Session.DeleteOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: session.ID}})
				mutex.Unlock()
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
