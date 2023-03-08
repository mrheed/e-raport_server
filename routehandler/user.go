package routehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/syahidnurrohim/restapi/database"
	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserData interface
type UserData struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username   string             `json:"username,omitempty" bson:"username,omitempty"`
	Password   string             `json:"password,omitempty" bson:"password,omitempty"`
	RoleKey    uint               `json:"role_key,omitempty" bson:"role_key,omitempty"`
	RoleString string             `json:"role,omitempty" bson:"role,omitempty"`
	LastLogin  interface{}        `json:"last_login" bson:"last_login"`
	Name       UserFName          `json:"name" bson:"name"`
}

// UserFName interface
type UserFName struct {
	Firstname string `json:"firstname" bson:"firstname"`
	Lastname  string `json:"lastname" bson:"lastname"`
}

// CreateUserController , this is handler method to the register event
func CreateUserController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user UserData
	json.NewDecoder(r.Body).Decode(&user)
	bytePass := []byte(user.Password)

	if user.RoleKey > 3 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": "Invalid Role"}`))
		return
	}

	result, err := database.User.InsertOne(context.Background(), bson.D{
		primitive.E{Key: "username", Value: user.Username},
		primitive.E{Key: "password", Value: tool.HashSaltPassword(bytePass)},
		primitive.E{Key: "name", Value: bson.D{primitive.E{Key: "firstname", Value: user.Name.Firstname}, primitive.E{Key: "lastname", Value: user.Name.Lastname}}},
		primitive.E{Key: "lastname", Value: user.Name.Lastname},
		primitive.E{Key: "role", Value: user.RoleString},
		primitive.E{Key: "role_key", Value: user.RoleKey},
		primitive.E{Key: "last_login", Value: time.Now()},
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message": ` + err.Error() + `}`))
		return
	}

	json.NewEncoder(w).Encode(result)
}

// GetUserController , send a user information with id parameter
func GetUserController(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var user UserData

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	result := db.User.FindOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&user)

	if result != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{message: ` + result.Error() + `}`))
		return
	}

	json.NewEncoder(w).Encode(user)

}

// EditUserController , handle user which will be change
func EditUserController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	var reqInput UserData
	var newData UserData

	_ = json.NewDecoder(r.Body).Decode(&reqInput)

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	result := db.User.FindOneAndUpdate(context.Background(),
		bson.D{primitive.E{Key: "_id", Value: id}}, bson.D{{Key: "$set",
			Value: bson.D{
				primitive.E{Key: "username", Value: reqInput.Username},
				primitive.E{Key: "password", Value: reqInput.Password},
			}}}).Decode(&newData)

	if result != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{message: ` + result.Error() + `}`))
		return
	}

	json.NewEncoder(w).Encode(newData)

}

// DeleteUserController , delete a user based on the object id
func DeleteUserController(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user UserData
	params := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(params["id"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{message :` + err.Error() + `}`))
		return
	}

	result := db.User.FindOneAndDelete(context.Background(), bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&user)

	if result != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{message :` + result.Error() + `}`))
	}

	json.NewEncoder(w).Encode(result)

}

// GetUsersController , get all users from the database and forward to the responses
func GetUsersController(w http.ResponseWriter, r *http.Request) {
	tool.HeadersHandler(w)

	var users []*UserData

	result, err := db.User.Find(context.Background(), bson.D{})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{message: ` + err.Error() + `}`))
		return
	}

	for result.Next(context.Background()) {
		var elem UserData
		err := result.Decode(&elem)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{message: ` + err.Error() + `}`))
			return
		}

		users = append(users, &elem)
	}

	json.NewEncoder(w).Encode(&users)
}
