package routehandler

import (
  "encoding/json"
  "net/http"

  mod "github.com/syahidnurrohim/restapi/models"
  tool "github.com/syahidnurrohim/restapi/utils"
)

func GetStudentsEkstraController(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {

  }
  ekstra := mod.NewEkstra()
  if kodeEkstra := tool.GQry("kode_ekstra", r); !tool.IsEmpty(kodeEkstra) {
    data, err := ekstra.GetStudentsEkstra(kodeEkstra)
    if err != nil {
      throw.Error(err.Error())
      return
    }
    throw.Response(&data)
  }
  if !tool.IsEmpty(tool.GQry("grade", r)) {
    data, err := ekstra.GetElevenStudents()
    if err != nil {
      throw.Error(err.Error())
      return
    }
    throw.Response(&data)
  }

}

func InsertStudentsEkstraController(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {

  }
  var body map[string]interface{}
  if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
    throw.Error(err.Error())
    return
  }
  if err := mod.NewEkstra().InsertStudentEkstra(body["kode_ekstra"].(string), body["data"].([]interface{})); err != nil {
    throw.Error(err.Error())
    return
  }
  throw.Response(tool.JSONGreen("data telah ditambahkan"))
}

func DeleteStudentsEkstraController(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {

  }
  var body map[string]interface{}
  if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
    throw.Error(err.Error())
    return
  }
  if err := mod.NewEkstra().DeleteStudentEkstra(body["filter"].(string), body["data"].([]interface{})); err != nil {
    throw.Error(err.Error())
    return
  }
  throw.Response(tool.JSONGreen("data telah dihapus"))
}
