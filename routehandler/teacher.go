package routehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	db "github.com/syahidnurrohim/restapi/database"
	mod "github.com/syahidnurrohim/restapi/models"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TeacherStruct struct {
	ID           primitive.ObjectID  `json:"_id,omitempty" bson:"_id,omitempty"`
	KelasDiampu  []tool.CustomSelect `json:"kelas_diampu" bson:"kelas_diampu"`
	JenisKelamin string              `json:"jeniskelamin" bson:"jeniskelamin"`
	IsWali       bool                `json:"is_wali" bson:"is_wali"`
	MapelDiampu  []tool.CustomSelect `json:"mapel" bson:"mapel"`
	Nama         string              `json:"nama" bson:"nama"`
	Wali         tool.CustomSelect   `json:"wali" bson:"wali"`
	NIP          string              `json:"nip" bson:"nip"`
}

type WaliType struct {
	KelasDiampu string `json:"kelas_diampu" bson:"kelas_diampu"`
	TahunAjaran int    `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

type InsertTeacher struct {
	Update mod.TeacherType `json:"update" bson:"update"`
	Filter tool.Filter     `json:"filter" bson:"filter"`
}

func GetTeachersController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(AllUser, r, w); err != nil {
	}
	teacher := mod.NewTeacher()
	data, err := teacher.GetAllTeacher()
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(&data)
}

func GetTeacherController(w http.ResponseWriter, r *http.Request) {
	if _, err := tool.VerifyHeader(AllUser, r, w); err == nil {
		var Teacher TeacherStruct
		params := mux.Vars(r)
		id, _ := primitive.ObjectIDFromHex(params["id"])
		err := db.Teacher.FindOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&Teacher)
		if err != nil {
			http.Error(w, tool.JSONErr("Records didn't found"), http.StatusMovedPermanently)
			return
		}
		json.NewEncoder(w).Encode(tool.JSONGreen(Teacher))
	}
}

func InsertTeacherController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(AllUser, r, w); err != nil {
	}
	var data []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		throw.Error(err.Error())
		return
	}
	teacher := mod.NewTeacher()
	if err := teacher.InsertTeacher(data); err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(tool.JSONGreen("data berhasil ditambahkan"))
}

func UpdateDeleteTeacherController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
	}
		teacher := mod.NewTeacher()
		if r.Method == "DELETE" {
			var dtc []primitive.M
			json.NewDecoder(r.Body).Decode(&dtc)
			cursor, err := db.Teacher.DeleteMany(context.Background(), bson.D{primitive.E{Key: "$or", Value: dtc}})
			if err != nil {
				http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
				return
			}
			deleteCount := strconv.FormatInt(cursor.DeletedCount, 10)
			json.NewEncoder(w).Encode(tool.JSONGreen(deleteCount + " records has deleted"))
		}
		if r.Method == "PUT" {
			var NewState map[string]interface{}
			json.NewDecoder(r.Body).Decode(&NewState)
			if err := teacher.UpdateTeacher(NewState["update"].(map[string]interface{}), NewState["filter"].(map[string]interface{})); err != nil {
				throw.Error(err.Error())
				return
			}
			throw.Response(tool.JSONGreen("data berhasil diupdate"))
		}
}
