package models

import (
  "context"
  "encoding/json"
  "fmt"
  db "github.com/syahidnurrohim/restapi/database"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "net/http"
  "strings"
)

type TeacherType struct {
  ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
  KelasDiampu  []StudentJurusan   `json:"kelas_diampu" bson:"kelas_diampu"`
  JenisKelamin string             `json:"jeniskelamin" bson:"jeniskelamin"`
  IsWali       bool               `json:"is_wali,omitempty" bson:"is_wali,omitempty"`
  MapelDiampu  []NamaMapel        `json:"mapel" bson:"mapel"`
  Nama         string             `json:"nama" bson:"nama"`
  Wali         StudentJurusan     `json:"wali" bson:"wali"`
  NIP          int                `json:"nip" bson:"nip"`
}

type TeacherConstruct struct {
  Writer http.ResponseWriter
}

func NewTeacher() *TeacherConstruct {
  return &TeacherConstruct{}
}

func (t *TeacherConstruct) GetAllTeacher() ([]bson.M, error) {
  var TeacerArray []bson.M

  cursor, err := db.Teacher.Find(context.Background(), bson.M{})
  if err != nil {
    return []bson.M{}, err
  }
  classOnSchoolYear, err := NewMisc().GetClassOnSchoolYear()
  if err != nil {
    return []bson.M{}, err
  }
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return []bson.M{}, err
  }

  for cursor.Next(context.Background()) {
    var Teacher TeacherType
    if err := cursor.Decode(&Teacher); err != nil {
      return []bson.M{}, err
    }
    voc := NewVocation()
    var kelasDiampu []bson.M
    var mapelDiampu []bson.M
    for _, d := range Teacher.KelasDiampu {
      jurusan, err := voc.GetSingleVocation(bson.M{"kode_kelas": d.KodeKelas, "tahun_ajaran": appSetting.TahunAjaran})
      if err != nil {
        continue
      }
      kelas := classOnSchoolYear.GradeOnSchoolYear[d.TahunAjaran]
      kelasDiampu = append(kelasDiampu, bson.M{"value": kelas + " " + jurusan.KodeKelas, "label": kelas + " " + jurusan.NamaKelas})
    }
    for _, d := range Teacher.MapelDiampu {
      mapel, err := NewMapel().GetSingleMapel(bson.M{"kode_mapel": d.KodeMapel, "tahun_ajaran": d.TahunAjaran})
      if err != nil {
        continue
      }
      mapelDiampu = append(mapelDiampu, bson.M{"value": mapel.KodeMapel, "label": mapel.NamaMapel})
    }
    kelas := classOnSchoolYear.GradeOnSchoolYear[Teacher.Wali.TahunAjaran]
    jurusan, _ := voc.GetSingleVocation(bson.M{"kode_kelas": Teacher.Wali.KodeKelas, "tahun_ajaran": appSetting.TahunAjaran})
    tmpData := bson.M{
      "_id":          Teacher.ID,
      "kelas_diampu": kelasDiampu,
      "jeniskelamin": Teacher.JenisKelamin,
      "is_wali":      Teacher.IsWali,
      "mapel":        mapelDiampu,
      "nama":         Teacher.Nama,
      "wali":         bson.M{"value": kelas + " " + jurusan.KodeKelas, "label": kelas + " " + jurusan.NamaKelas},
      "nip":          Teacher.NIP,
    }
    TeacerArray = append(TeacerArray, tmpData)
  }
  return TeacerArray, nil
}

func (t *TeacherConstruct) VerifyStruct(data bson.M) (TeacherType, error) {
  var result TeacherType
  byteData, err := json.Marshal(data)
  if err != nil {
    return TeacherType{}, err
  }
  if err := json.Unmarshal(byteData, &result); err != nil {
    return TeacherType{}, err
  }
  return result, nil
}

func (t *TeacherConstruct) InsertTeacher(data []map[string]interface{}) error {
  var insertData []interface{}
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  classOnSchoolYear, err := NewMisc().GetClassOnSchoolYear()
  if err != nil {
    return err
  }
  for _, d := range data {
    for _, kd := range d["kelas_diampu"].([]interface{}) {
      kd, ok := kd.(map[string]interface{})
      if !ok {
        continue
      }
      splitted := strings.Split(kd["value"].(string), " ")
      kd["kode_kelas"], kd["tahun_ajaran"] = splitted[1], classOnSchoolYear.SchoolYearOnGrade[splitted[0]]
    }
    for _, md := range d["mapel"].([]interface{}) {
      md, ok := md.(map[string]interface{})
      if !ok {
        continue
      }
      md["kode_mapel"], md["tahun_ajaran"] = md["value"], appSetting.TahunAjaran
    }
    wali, ok := d["wali"].(map[string]interface{})
    if ok {
      if wali["value"] == "bukan_wali" {
        wali["kode_kelas"], wali["tahun_ajaran"] = "", 0
      } else {
        splitted := strings.Split(wali["value"].(string), " ")
        wali["kode_kelas"], wali["tahun_ajaran"], d["is_wali"] = splitted[1], classOnSchoolYear.SchoolYearOnGrade[splitted[0]], true
      }
    } else {
      d["wali"] = map[string]interface{}{"kode_kelas": "", "tahun_ajaran": 0}
    }
    verified, err := t.VerifyStruct(d)
    if err != nil {
      return err
    }
    insertData = append(insertData, verified)
  }
  fmt.Printf("%+v\n", insertData)
  _, err = db.Teacher.InsertMany(context.Background(), insertData)
  if err != nil {
    return err
  }
  return nil
}

func (t *TeacherConstruct) UpdateTeacher(data map[string]interface{}, filter map[string]interface{}) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  classOnSchoolYear, err := NewMisc().GetClassOnSchoolYear()
  if err != nil {
    return err
  }
  for _, d := range data["kelas_diampu"].([]interface{}) {
    d := d.(map[string]interface{})
    splitted := strings.Split(d["value"].(string), " ")
    d["kode_kelas"], d["tahun_ajaran"] = splitted[1], classOnSchoolYear.SchoolYearOnGrade[splitted[0]]
  }
  for _, d := range data["mapel"].([]interface{}) {
    d := d.(map[string]interface{})
    d["kode_mapel"], d["tahun_ajaran"] = d["value"], appSetting.TahunAjaran
  }
  wali := data["wali"].(map[string]interface{})
  splitted := strings.Split(wali["value"].(string), " ")
  wali["kode_kelas"], wali["tahun_ajaran"] = splitted[1], classOnSchoolYear.SchoolYearOnGrade[splitted[0]]
  ID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
  if err != nil {
    return err
  }
  verified, err := t.VerifyStruct(data)
  if err != nil {
    return err
  }
  _, err = db.Teacher.UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": verified})
  fmt.Printf("%+v\n", data)
  return nil
}
