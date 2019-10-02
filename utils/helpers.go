package utils

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// JGType struct
type JGType struct {
	Status  string      `json:"status"`
	Message interface{} `json:"message"`
}

type CustomSelect struct {
	Label string      `json:"label" bson:"label"`
	Value interface{} `json:"value" bson:"value"`
}

type Role struct {
	AllUser   []int
	WaliKelas []int
	Admin     []int
	Guru      []int
	GuruAgama []int
	GuruPKN   []int
}

// InArray , check values from array sliice
func InArray(value string, array []string) (isExists bool, length int, index []int) {
	isExists, length, index = false, 0, []int{}

	for i, val := range array {
		if val == value {
			index = append(index, i)
			isExists = true
			length++
		}
	}
	return isExists, length, index
}

// ReverseArr helpers
func ReverseArr(input []string) []string {
	if len(input) == 0 {
		return input
	}
	return append(ReverseArr(input[1:]), input[0])
}

// VerifyHeader func
func VerifyHeader(role []int, r *http.Request, w http.ResponseWriter) (string, error) {
	authToken := r.Header.Get(Header)
	splittedToken := ReverseArr(strings.Split(authToken, " "))[0]
	if splittedToken == "" {
		return "", errors.New("auth token isn't available")
	}
	if err := CheckAuthToken(splittedToken, role, r); err != nil {
		return "", err
	}
	return splittedToken, nil
}

// JSONErr func
func JSONErr(errorMessage string) string {
	return `{"status": "error","message": "` + errorMessage + `"}`
}

// JSONGreen func
func JSONGreen(responseMessage interface{}) JGType {
	return JGType{
		Status:  "success",
		Message: responseMessage,
	}
}

// EmptyExecuted cehck if empty data is known
func EmptyExecuted(v interface{}) error {
	numField := reflect.ValueOf(v).NumField()
	for i := 0; i < numField; i++ {
		if IsEmpty(v) {
			return errors.New("Mohon mengisi " + GSName(v, i))
		}
	}
	return nil
}

// Contains check value inside a slice
func Contains(slice []interface{}, val interface{}) bool {
	for _, s := range slice {
		return s == val
	}
	return false
}

// IsEmpty check emptiness on a struct
func IsEmpty(v interface{}) bool {
	switch v.(type) {
	case string:
		return v == ""
	case int, float32, float64:
		return v == -1
	default:
		return false
	}
}

// MapStruct convert struct into map
func MapStruct(v interface{}) map[string]interface{} {
	numField := reflect.ValueOf(v).NumField()
	values := make(map[string]interface{}, numField)
	for i := 0; i < numField; i++ {
		name := reflect.Indirect(reflect.ValueOf(v)).Type().Field(i).Name
		values[name] = reflect.ValueOf(v).Field(i).Interface()
	}
	return values
}

// GSName get struct key name
func GSName(v interface{}, i int) string {
	return reflect.Indirect(reflect.ValueOf(v)).Type().Field(i).Name
}

func GetRole() Role {
	var Roles Role
	path, _ := filepath.Abs("./utils/Role.json")
	plan, _ := ioutil.ReadFile(path)
	json.Unmarshal(plan, &Roles)
	return Roles
}

func AgURI(filename string) string {
	return "./json/" + filename + ".json"
}

func GQry(key string, r *http.Request) string {
	return r.URL.Query().Get(key)
}

func ProcessDataAggregate(result *primitive.A, w http.ResponseWriter, collection *mongo.Collection, pipe []bson.M) bool {
	cursor, err := collection.Aggregate(context.Background(), pipe)
	if err != nil {
		http.Error(w, JSONErr(err.Error()), http.StatusInternalServerError)
		return false
	}
	for cursor.Next(context.Background()) {
		var elem bson.D
		cursor.Decode(&elem)
		*result = append(*result, elem.Map())
	}
	return true
}

func ProcessPipeMiddleware(pipe *[]bson.M, secondPipe string) {
	prependQ := []bson.M{}
	err := json.Unmarshal([]byte(secondPipe), &prependQ)
	log.Println(err)
	*pipe = append(prependQ, *pipe...)
}

func ProcessPipeAggregate(location string, pipe *[]bson.M, w http.ResponseWriter) {
	plan, _ := LocateReadFile(location)
	json.Unmarshal(plan, &pipe)
}

func LocateReadFile(location string) ([]byte, error) {
	path, err := filepath.Abs(location)
	if err != nil {
		return nil, err
	}
	plan, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return plan, nil
}
