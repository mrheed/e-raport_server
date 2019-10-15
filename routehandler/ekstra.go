package routehandler

import (
	"encoding/json"
	"net/http"

	mod "github.com/syahidnurrohim/restapi/models"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func GetEkstraController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {

	}
	ekstra := mod.NewEkstra()
	data, err := ekstra.GetEkstraData()
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(&data)
}

func InsertEkstraController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {

	}
	var body []mod.EkstraStruct
	ekstra := mod.NewEkstra()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		throw.Error(err.Error())
		return
	}
	if err := ekstra.InsertEkstraData(body); err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(tool.JSONGreen("data berhasil ditambahkan"))
}

func UpdateDeleteEkstraController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {

	}
	ekstra := mod.NewEkstra()
	if r.Method == "PUT" {
		var body bson.M
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := ekstra.UpdateEkstraData(body["filter"].(map[string]interface{}), body["update"].(map[string]interface{})); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil diubah"))
	} else if r.Method == "DELETE" {
		var body []bson.M
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := ekstra.DeleteEkstraData(body); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil dihapus"))
	}
}
