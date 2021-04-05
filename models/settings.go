package models

import (
  "context"
  db "github.com/syahidnurrohim/restapi/database"
  "go.mongodb.org/mongo-driver/bson"
)

type Application struct {
  TahunAjaran      int    `json:"tahun_ajaran,omitempty" bson:"tahun_ajaran,omitempty"`
  NamaSekolah      string `json:"nama_sekolah,omitempty" bson:"nama_sekolah,omitempty"`
  DeskripsiSekolah string `json:"deskripsi_sekolah,omitempty" bson:"deskripsi_sekolah,omitempty"`
  Semester         int    `json:"semester,omitempty" bson:"semester,omitempty"`
}

func NewSetting() *Application {
  return &Application{}
}

func (a *Application) GetAppSetting() (Application, error) {
  var AppResult Application
  err := db.AppSetting.FindOne(context.Background(), bson.D{}).Decode(&AppResult)
  if err != nil {
    return Application{}, err
  }
  return AppResult, nil
}
