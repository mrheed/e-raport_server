package models

import (
  "context"
  "encoding/json"
  "fmt"
  "net/http"
  "strconv"
  "strings"

  "go.mongodb.org/mongo-driver/mongo"

  db "github.com/syahidnurrohim/restapi/database"
  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportResult struct {
  Data           DataReport `json:"data"`
  KKM            int        `json:"kkm"`
  Mapel          string     `json:"mapel"`
  Materi         string     `json:"materi"`
  RataRata       int        `json:"rata2"`
  TanggalUlangan string     `json:"tanggal_ulangan"`
  Terendah       int        `json:"terendah"`
  Tertinggi      int        `json:"tertinggi"`
}

type DataReport struct {
  Nama         string `json:"nama"`
  NilaiUlangan int    `json:"nilai_ulangan"`
  Remidi       bool   `json:"remidi"`
  TelahRemidi  bool   `json:"telah_remidi"`
  NilaiRemidi  int    `json:"nilai_remidi"`
}

type Reporter struct {
  Result *primitive.A
  Pipe   *[]bson.M
  Writer http.ResponseWriter
}

func NewReporter(pipe *[]bson.M, result *primitive.A, w http.ResponseWriter) *Reporter {
  return &Reporter{
    Result: result,
    Pipe:   pipe,
    Writer: w,
  }
}

func (r *Reporter) GetReportStudent(jurusan string, tahunMasuk int) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  cursor, err := db.Student.Find(context.Background(), bson.M{"jurusan.kode_kelas": jurusan, "tahun_masuk": tahunMasuk, "jurusan.tahun_ajaran": appSetting.TahunAjaran})
  if err != nil {
    return err
  }
  for cursor.Next(context.Background()) {
    var Student Student
    if err := cursor.Decode(&Student); err != nil {
      return err
    }
    coded := map[string]interface{}{
      "label": Student.Nama,
      "value": Student.NIS,
    }
    *r.Result = append(*r.Result, coded)
  }
  return nil
}

func (r *Reporter) GetReportSubject(jurusan string, tahunAjaran int) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  cursor, err := db.Mapel.Find(context.Background(), bson.M{"$and": []bson.M{
    bson.M{"$or": []bson.M{
      bson.M{"mapel_kelas": bson.M{"$elemMatch": bson.M{"jurusan": jurusan, "tahun_ajaran": tahunAjaran}}},
      bson.M{"mapel_kelas": bson.M{"$elemMatch": bson.M{"jurusan": "semua kelas", "tahun_ajaran": tahunAjaran}}},
    }},
    bson.M{"tahun_ajaran": appSetting.TahunAjaran},
  }})
  if err != nil {
    return err
  }
  for cursor.Next(context.Background()) {
    var Mapel MapelStruct
    if err := cursor.Decode(&Mapel); err != nil {
      return err
    }
    coded := map[string]string{
      "value": Mapel.KodeMapel,
      "label": Mapel.NamaMapel,
    }
    *r.Result = append(*r.Result, coded)
  }
  return nil
}

func (r *Reporter) GetReportMaterial(mapel string) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  cursor, err := db.Kompetensi.Find(context.Background(), bson.M{"nama_mapel.kode_mapel": mapel, "tahun_ajaran": appSetting.TahunAjaran})
  if err != nil {
    return err
  }
  for cursor.Next(context.Background()) {
    var Material MaterialStruct
    if err := cursor.Decode(&Material); err != nil {
      return err
    }
    coded := map[string]string{
      "value": Material.KodeMateri,
      "label": Material.NamaMateri,
    }
    *r.Result = append(*r.Result, coded)
  }
  return nil
}

func (r *Reporter) GetReportTaskName(mapel string) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  distinc, err := db.NilaiTugas.Distinct(context.Background(), "nama_tugas", bson.M{"mapel": mapel, "tahun_ajaran": appSetting.TahunAjaran})
  if err != nil {
    return err
  }
  for _, s := range distinc {
    coded := map[string]interface{}{
      "value": s,
      "label": s,
    }
    *r.Result = append(*r.Result, coded)
  }
  return nil
}

func (r *Reporter) GetReportExamResult(nis string, materi string, mapel string, jurusan string, tipe string) error {
  nis = splitNisToMap(nis)
  col := jkmk(tipe == "UH", db.NilaiUH, db.NilaiPTPAS)
  fmt.Printf("%+v\n", col, tipe)
  pipe2 := `[
  {"$lookup": {
    "from": "c_application_setting",
    "pipeline": [],
    "as": "setting"
  }},
  {"$unwind": "$setting"},
  {"$match": {"$and": [
  {"$or": ` + nis + `},
  ` + jkmk1(tipe == "UH", `{"materi": "`+materi+`"},`, ``) + `
  ` + jkmk1(tipe != "UH", `{"tipe": "`+tipe+`"},`, ``) + `
  {"mapel": "` + mapel + `"},
  {"$expr": {"$eq": ["$semester", "$setting.semester"]}},
  {"$expr": {"$eq": ["$tahun_ajaran", "$setting.tahun_ajaran"]}}
  ]}}
  ` + jkmk1(tipe == "UH", `
  ,{"$lookup": {
    "from": "c_materi",
    "let": {"kode_materi": "$materi", "tahun_ajaran": "$setting.tahun_ajaran"},
    "pipeline": [
    {"$match": {"$and": 
    [
    {"$expr": {"$eq": ["$kode_materi", "$$kode_materi"]}},
    {"$expr": {"$eq": ["$tahun_ajaran", "$$tahun_ajaran"]}}
    ]
  }}
  ],
  "as": "materi"
}},
{"$unwind": "$materi"}`, ``) + `
]`
tool.ProcessPipeAggregate(tool.AgURI("ExamPrint"), r.Pipe, r.Writer)
tool.ProcessPipeMiddleware(r.Pipe, pipe2)
fmt.Printf("%+v\n", r.Pipe)
cursor, err := col.Aggregate(context.Background(), *r.Pipe)
if err != nil {
  return err
}
for cursor.Next(context.Background()) {
  var elem bson.M
  cursor.Decode(&elem)
  elem["kelas"] = jurusan
  *r.Result = append(*r.Result, elem)
}
return nil
}

func (r *Reporter) GetReportTaskResult(nis string, namaTugas string, mapel string, jurusan string) error {
  nis = splitNisToMap(nis)
  pipe2 := `[
  {"$lookup": {
    "from": "c_application_setting",
    "pipeline": [],
    "as": "setting"
  }},
  {"$unwind": "$setting"},
  {"$match": {"$and": [
  {"$or": ` + nis + `},
  {"nama_tugas": "` + namaTugas + `"},
  {"mapel": "` + mapel + `"},
  {"$expr": {"$eq": ["$setting.semester", "$semester"]}},
  {"$expr": {"$eq": ["$setting.tahun_ajaran", "$tahun_ajaran"]}}
  ]}}
  ]`
  tool.ProcessPipeAggregate(tool.AgURI("TaskPrint"), r.Pipe, r.Writer)
  tool.ProcessPipeMiddleware(r.Pipe, pipe2)
  cursor, err := db.NilaiTugas.Aggregate(context.Background(), *r.Pipe)
  if err != nil {
    return err
  }
  for cursor.Next(context.Background()) {
    var elem bson.M
    cursor.Decode(&elem)
    elem["kelas"] = jurusan
    *r.Result = append(*r.Result, elem)
  }
  return nil
}

func splitNisToMap(nis string) string {
  var mappedNis []map[string]int
  for _, s := range strings.Split(nis, ",") {
    intNis, _ := strconv.Atoi(s)
    mappedNis = append(mappedNis, map[string]int{"nis": intNis})
  }
  wa, _ := json.Marshal(mappedNis)
  return string(wa)
}

func jkmk(con1 bool, result1 *mongo.Collection, result2 *mongo.Collection) *mongo.Collection {
  if con1 {
    return result1
  }
  return result2
}

func jkmk1(con1 bool, result1 string, result2 string) string {
  if con1 {
    return result1
  }
  return result2
}
