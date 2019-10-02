package routehandler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	mod "github.com/syahidnurrohim/restapi/models"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type KelasStruct struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	KodeKelas string             `json:"kode_kelas,omitempty" bson:"kode_kelas,omitempty"`
	NamaKelas string             `json:"nama_kelas,omitempty" bson:"nama_kelas,omitempty"`
}

type KelasFilter struct {
	Update KelasStruct `json:"update" bson:"update"`
	Filter tool.Filter `json:"filter" bson:"filter"`
}

func GetClassesController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
	}
	jurusan := mod.NewVocation()
	data, err := jurusan.GetAllVocation()
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(&data)
}

func GetClassController(w http.ResponseWriter, r *http.Request) {
	tool.HeadersHandler(w)
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err == nil {
		params := mux.Vars(r)["misc"]
		if params == "grade" {
			var data []interface{}
			plan, err := locateReadFile("'./json/Grade.json")
			if err != nil {
				http.Error(w, tool.JSONErr(err.Error()), http.StatusInternalServerError)
				return
			}
			err = json.Unmarshal(plan, &data)
			if err != nil {
				http.Error(w, tool.JSONErr(err.Error()), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(&data)
		}
	}
}

func InsertClassController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(Admin, r, w); err != nil {
	}
	jurusan := mod.NewVocation()
	err := jurusan.InsertVocation(r)
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(tool.JSONGreen("data berhasil ditambahkan"))
}

func UpdateDeleteClassController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(Admin, r, w); err != nil {
	}
	jurusan := mod.NewVocation()
	if r.Method == "DELETE" {
		var dtc []bson.M
		if err := json.NewDecoder(r.Body).Decode(&dtc); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := jurusan.DeleteVocation(dtc); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil di hapus"))
	}
	if r.Method == "PUT" {
		var decoded map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := jurusan.UpdateVocation(decoded["filter"].(map[string]interface{}), decoded["update"].(map[string]interface{})); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil diupdate"))
	}
}
