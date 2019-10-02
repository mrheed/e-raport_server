package routehandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	db "github.com/syahidnurrohim/restapi/database"
	mod "github.com/syahidnurrohim/restapi/models"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KompetensiStruct struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	KodeMateri string             `json:"kode_materi,omitempty" bson:"kode_materi,omitempty"`
	NamaMapel  tool.CustomSelect  `json:"nama_mapel,omitempty" bson:"nama_mapel,omitempty"`
	NamaMateri string             `json:"nama_materi,omitempty" bson:"nama_materi,omitempty"`
}

type KompetensiFilter struct {
	Update mod.MaterialStruct `json:"update" bson:"update"`
	Filter tool.Filter        `json:"filter" bson:"filter"`
}

func GetCompetenciesController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
		throw.Error(err.Error())
		return
	}
	materi := mod.NewMateri()
	data, err := materi.GetAllMaterial()
	if err != nil {
		throw.Error(err.Error())
	}
	throw.Response(&data)
}

func GetCompetenceController(w http.ResponseWriter, r *http.Request) {
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err == nil {
		var Kompetensi mod.MaterialStruct
		params := mux.Vars(r)["id"]
		err := db.Kompetensi.FindOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: params}}).Decode(&Kompetensi)
		if err != nil {
			http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
			return
		}
		json.NewEncoder(w).Encode(&Kompetensi)
	}
}

func InsertCompetenceController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
	}
	materi := mod.NewMateri()
	err := materi.InsertMaterial(r)
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(tool.JSONGreen("data telah ditambahkan"))
}

func UpdateDeleteCompetenceController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
		throw.Error(err.Error())
		return
	}
	materi := mod.NewMateri()
	if r.Method == "DELETE" {
		var decoded []map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := materi.DeleteMaterial(decoded); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil dihapus"))
	}
	if r.Method == "PUT" {
		var decoded map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := materi.UpdateMaterial(decoded["filter"].(map[string]interface{}), decoded["update"].(map[string]interface{})); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil diubah"))
	}
}
