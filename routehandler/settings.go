package routehandler

import (
  "context"
  "encoding/json"
  "log"
  "net/http"

  "go.mongodb.org/mongo-driver/bson"

  "github.com/gorilla/mux"

  db "github.com/syahidnurrohim/restapi/database"
  tool "github.com/syahidnurrohim/restapi/utils"
)

type Application struct {
  TahunAjaran      int    `json:"tahun_ajaran,omitempty" bson:"tahun_ajaran,omitempty"`
  NamaSekolah      string `json:"nama_sekolah,omitempty" bson:"nama_sekolah,omitempty"`
  DeskripsiSekolah string `json:"deskripsi_sekolah,omitempty" bson:"deskripsi_sekolah,omitempty"`
  Semester         int    `json:"semester,omitempty" bson:"semester,omitempty"`
}

func UpdateSettingController(w http.ResponseWriter, r *http.Request) {
  params := mux.Vars(r)
  if params["tab"] == "application" {
    var AppData Application
    err := json.NewDecoder(r.Body).Decode(&AppData)
    if err != nil {
      http.Error(w, tool.JSONErr("Error saat memproses data"), http.StatusMovedPermanently)
      return
    }
    if AppData.TahunAjaran == 0 {
      http.Error(w, tool.JSONErr("Mohon mengisi tahun ajaran"), http.StatusMovedPermanently)
      return
    }
    if AppData.Semester == 0 {
      http.Error(w, tool.JSONErr("Mohon mengisi semester"), http.StatusMovedPermanently)
      return
    }
    if AppData.DeskripsiSekolah == "" {
      http.Error(w, tool.JSONErr("Mohon mengisi deskripsi sekolah"), http.StatusMovedPermanently)
      return
    }
    if AppData.NamaSekolah == "" {
      http.Error(w, tool.JSONErr("Mohon mengisi nama sekolah"), http.StatusMovedPermanently)
      return
    }
    _, err = db.AppSetting.UpdateOne(context.Background(), bson.D{}, bson.M{"$set": AppData})

    if err != nil {
      http.Error(w, tool.JSONErr("Tidak dapat mengubah data"), http.StatusMovedPermanently)
      return
    }
    json.NewEncoder(w).Encode(tool.JSONGreen("Data berhasil di ubah"))
  }
}

func GetSettingController(w http.ResponseWriter, r *http.Request) {
  if gQry("type", r) == "application" {
    var AppData Application
    err := db.AppSetting.FindOne(context.Background(), bson.D{}).Decode(&AppData)
    if err != nil {
      log.Println(err.Error())
      http.Error(w, tool.JSONErr("Data tidak tersedia, mohon konfigurasi pengaturan aplikasi dari awal"), http.StatusMovedPermanently)
      return
    }
    json.NewEncoder(w).Encode(bson.M{"application": AppData})
  }
}

func InsertSettingController(w http.ResponseWriter, r *http.Request) {
  tab := mux.Vars(r)["tab"]
  if tab == "application" {
    var AppData Application
    err := json.NewDecoder(r.Body).Decode(&AppData)
    if err != nil {
      http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
      return
    }
    if AppData.TahunAjaran == 0 {
      http.Error(w, tool.JSONErr("Mohon mengisi tahun ajaran"), http.StatusMovedPermanently)
      return
    }
    if AppData.Semester == 0 {
      http.Error(w, tool.JSONErr("Mohon mengisi semester"), http.StatusMovedPermanently)
      return
    }
    if AppData.DeskripsiSekolah == "" {
      http.Error(w, tool.JSONErr("Mohon mengisi deskripsi sekolah"), http.StatusMovedPermanently)
      return
    }
    if AppData.NamaSekolah == "" {
      http.Error(w, tool.JSONErr("Mohon mengisi nama sekolah"), http.StatusMovedPermanently)
      return
    }
    _, err = db.AppSetting.InsertOne(context.Background(), AppData)
    if err != nil {
      http.Error(w, tool.JSONErr("Tidak dapat menyimpan data"), http.StatusMovedPermanently)
      return
    }
    json.NewEncoder(w).Encode(tool.JSONGreen("Data berhasil tersimpan"))
  }
}
