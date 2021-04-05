package routehandler

import (
  "context"
  "encoding/json"
  "github.com/gorilla/mux"
  "net/http"

  db "github.com/syahidnurrohim/restapi/database"
  mod "github.com/syahidnurrohim/restapi/models"
  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

type MapelFilter struct {
  Update mod.MapelStruct `json:"update" bson:"update"`
  Filter tool.Filter     `json:"filter" bson:"filter"`
}

func GetSubjectsController(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
    //throw.Error(err.Error())
    //return
  }
  mapel := mod.NewMapel()
  if jurusan := tool.GQry("jurusan", r); !tool.IsEmpty(jurusan) {
    data, err := mapel.GetRestructuredMapelWithJurusan(jurusan)
    if err != nil {
      throw.Error(err.Error())
      return
    }
    throw.Response(&data)
  }
  data, err := mapel.GetAllMapel()
  if err != nil {
    throw.Error(err.Error())
    return
  }
  throw.Response(&data)
}

func GetSubjectController(w http.ResponseWriter, r *http.Request) {
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
  }
  var Mapel mod.MapelStruct
  params := mux.Vars(r)["id"]
  err := db.Mapel.FindOne(context.Background(), bson.D{primitive.E{Key: "_id", Value: params}}).Decode(&Mapel)
  if err != nil {
    http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
    return
  }
  json.NewEncoder(w).Encode(&Mapel)
}

func InsertSubjectController(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  //if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
  //	throw.Error(err.Error())
  //	return
  //}
  var decoded []bson.M
  if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
    throw.Error(err.Error())
    return
  }
  if err := mod.NewMapel().InsertMapel(decoded); err != nil {
    throw.Error(err.Error())
    return
  }
  throw.Response(tool.JSONGreen("data telah ditambahkan"))
}

func UpdateDeleteSubjectController(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
    throw.Error(err.Error())
    return
  }
  mapel := mod.NewMapel()
  if r.Method == "DELETE" {
    if err := mapel.DeleteMapel(r); err != nil {
      throw.Error(err.Error())
      return
    }
    throw.Response(tool.JSONGreen("data berhasil dihapus"))
  }
  if r.Method == "PUT" {
    if err := mapel.UpdateMapel(r); err != nil {
      throw.Error(err.Error())
      return
    }
    throw.Response(tool.JSONGreen("data berhasil diubah"))
  }
}
