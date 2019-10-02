package routehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	db "github.com/syahidnurrohim/restapi/database"
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
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	Username   string             `json:"username" bson:"username"`
	Password   string             `json:"password" bson:"password"`
	RoleKey    int                `json:"role_key,omitempty" bson:"role_key,omitempty"`
	RoleString string             `json:"role,omitempty" bson:"role,omitempty"`
}

// SignInRequestHandler method
func SignInRequestHandler(w http.ResponseWriter, r *http.Request) {

	var cridentials SignInCridentials
	var userCridentials UserData
	var plainPassword string

	err := json.NewDecoder(r.Body).Decode(&cridentials)

	plainPassword = cridentials.Password

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{message : ` + err.Error() + `}`))
		return
	}

	resultErr := db.User.FindOne(context.Background(), bson.D{
		primitive.E{Key: "username", Value: cridentials.Username}}).Decode(&cridentials)

	if tool.ComparePassword([]byte(plainPassword), cridentials.Password) {

		resultSign := db.User.FindOne(context.Background(), bson.D{
			primitive.E{Key: "username", Value: cridentials.Username},
			primitive.E{Key: "password", Value: cridentials.Password},
			primitive.E{Key: "_id", Value: cridentials.ID}}).Decode(&cridentials)

		if resultSign != nil {
			http.Error(w, `{"message" : "Unknown user"}`, http.StatusUnauthorized)
			return
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		db.Session.InsertOne(context.Background(), bson.D{
			primitive.E{Key: "_id", Value: primitive.NewObjectID()},
			primitive.E{Key: "user_id", Value: cridentials.ID},
			primitive.E{Key: "username", Value: cridentials.Username},
			primitive.E{Key: "role", Value: cridentials.RoleString},
			primitive.E{Key: "role_key", Value: cridentials.RoleKey},
			primitive.E{Key: "token", Value: tokenString},
			primitive.E{Key: "date_added", Value: time.Now()},
			primitive.E{Key: "date_expired", Value: expirationTime},
		})

		db.User.UpdateOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: cridentials.ID}}, bson.D{{
			Key: "$set",
			Value: bson.D{
				primitive.E{
					Key:   "last_login",
					Value: time.Now(),
				}}}})

		db.User.FindOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: cridentials.ID}}).Decode(&userCridentials)

		payload := map[string]interface{}{
			"user_data": &UserData{
				ID:       cridentials.ID,
				Username: cridentials.Username,
				Name: UserFName{
					Firstname: userCridentials.Name.Firstname,
					Lastname:  userCridentials.Name.Lastname,
				},
				RoleString: userCridentials.RoleString,
				RoleKey:    userCridentials.RoleKey,
				LastLogin:  userCridentials.LastLogin,
			},
			"http_data": &http.Cookie{
				Name:    "token",
				Value:   tokenString,
				Expires: expirationTime,
			},
		}

		json.NewEncoder(w).Encode(payload)

	} else {
		http.Error(w, `{"message" : "Unknown user"}`, http.StatusUnauthorized)
		return
	}

	if resultErr != nil {
		http.Error(w, `{"message" : "`+resultErr.Error()+`"}`, http.StatusUnauthorized)
		return
	}

}

// PurgeSingleToken , remove single session from database
func PurgeSingleToken(w http.ResponseWriter, r *http.Request) {

	if splittedToken, err := tool.VerifyHeader(AllUser, r, w); err == nil {
		var DeletedSession SessionStruct

		db.Session.FindOneAndDelete(context.Background(), bson.D{
			primitive.E{Key: "token", Value: splittedToken},
		}).Decode(&DeletedSession)
	}

}
