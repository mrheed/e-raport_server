package models

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  db "github.com/syahidnurrohim/restapi/database"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "net/http"
)

type Jurusan struct {
  ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
  KodeKelas   string             `json:"kode_kelas,omitempty" bson:"kode_kelas,omitempty"`
  NamaKelas   string             `json:"nama_kelas,omitempty" bson:"nama_kelas,omitempty"`
  TahunAjaran int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

func NewVocation() *Jurusan {
  return &Jurusan{}
}

func (j *Jurusan) GetSingleVocation(filter bson.M) (Jurusan, error) {
  var result Jurusan
  err := db.Jurusan.FindOne(context.Background(), filter).Decode(&result)
  if err != nil {
    return Jurusan{}, err
  }
  return result, nil
}

func (j *Jurusan) GetAllVocation() ([]Jurusan, error) {
  var result []Jurusan
  setting := NewSetting()
  appSetting, err := setting.GetAppSetting()
  if err != nil {
    return []Jurusan{}, err
  }
  cursor, err := db.Jurusan.Find(context.Background(), bson.M{"tahun_ajaran": appSetting.TahunAjaran})
  if err != nil {
    return []Jurusan{}, err
  }
  for cursor.Next(context.Background()) {
    var tmpData Jurusan
    if err := cursor.Decode(&tmpData); err != nil {
      return []Jurusan{}, err
    }
    result = append(result, tmpData)
  }
  return result, nil
}

func (j *Jurusan) InsertVocation(r *http.Request) error {
  var structData []interface{}
  if err := json.NewDecoder(r.Body).Decode(&structData); err != nil {
    return err
  }
  setting := NewSetting()
  appSetting, err := setting.GetAppSetting()
  if err != nil {
    return err
  }
  for _, d := range structData {
    tmpData, passed := d.(map[string]interface{})
    if !passed {
      return errors.New("error type assertion")
    }
    tmpData["_id"] = primitive.NewObjectID()
    tmpData["tahun_ajaran"] = appSetting.TahunAjaran
    exist := findAndExist(db.Jurusan, bson.M{"kode_kelas": tmpData["kode_kelas"], "tahun_ajaran": tmpData["tahun_ajaran"]})
    if exist {
      return errors.New("error: kelas " + tmpData["nama_kelas"].(string) + " sudah tersedia")
    }
  }
  fmt.Printf("%+v\n", structData)
  _, err = db.Jurusan.InsertMany(context.Background(), structData)
  if err != nil {
    return err
  }
  return nil
}

func (j *Jurusan) UpdateVocation(filter map[string]interface{}, update map[string]interface{}) error {
  ID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
  if err != nil {
    return err
  }
  delete(update, "_id")
  _, err = db.Jurusan.UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": update})
  if err != nil {
    return err
  }
  return nil
}

func (j *Jurusan) DeleteVocation(filter []bson.M) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  for _, d := range filter {
    d["tahun_ajaran"] = appSetting.TahunAjaran
    _, err := db.Jurusan.DeleteOne(context.Background(), d)
    if err != nil {

    }
  }
  fmt.Printf("%+v\n", filter)
  return nil
}
