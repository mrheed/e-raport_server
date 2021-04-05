package models

import (
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "strings"

  db "github.com/syahidnurrohim/restapi/database"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

type EkstraStruct struct {
  ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
  KodeEkstra  string             `json:"kode_ekstra,omitempty" bson:"kode_ekstra,omitempty"`
  NamaEkstra  string             `json:"nama_ekstra" bson:"nama_ekstra"`
  Pelatih     string             `json:"pelatih" bson:"pelatih"`
  TahunAjaran int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

type StudentEkstra struct {
  NIS   int    `json:"nis" bson:"nis"`
  Nama  string `json:"nama" bson:"nama"`
  Kelas string `json:"kelas" bson:"kelas"`
}

type ElevenStudents struct {
  ID          primitive.ObjectID `json:"_id" bson:"_id"`
  KodeEkstra  string             `json:"kode_ekstra" bson:"kode_ekstra"`
  TahunAjaran int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
  NIS         []int              `json:"nis" bson:"nis"`
}

type Siswa struct {
  Value int    `json:"value" bson:"value"`
  Label string `json:"label" bson:"label"`
}

func NewEkstra() *EkstraStruct {
  return &EkstraStruct{}
}

func (e *EkstraStruct) GetEkstraData() ([]EkstraStruct, error) {
  var result []EkstraStruct
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return []EkstraStruct{}, err
  }
  cursor, err := db.Ekstra.Find(context.Background(), bson.M{"tahun_ajaran": appSetting.TahunAjaran})
  if err != nil {
    return []EkstraStruct{}, err
  }
  for cursor.Next(context.Background()) {
    var tmp EkstraStruct
    if err := cursor.Decode(&tmp); err != nil {
      continue
    }
    result = append(result, tmp)
  }
  return result, nil
}

func (e *EkstraStruct) GetStudentsEkstra(kode_ekstra string) (bson.M, error) {
  var result bson.M
  var pipe []bson.M
  plan := `
  [{"$lookup": {
    "from": "c_application_setting",
    "pipeline": [
    {"$project": {
      "kelas_x": "$tahun_ajaran",
      "kelas_xi": {"$sum": ["$tahun_ajaran", -1]},
      "kelas_xii": {"$sum": ["$tahun_ajaran", -2]}
    }}
    ],
    "as": "setting"
  }},
  {"$unwind": "$setting"},
  {"$match": {"$and": [
  {"kode_ekstra": "` + kode_ekstra + `"},
  {"$expr": {"$eq": ["$setting.kelas_x", "$tahun_ajaran"]}}
  ]
}},
{"$lookup": {
  "from": "c_students",
  "let": {"daftar_nis": "$nis"},
  "pipeline": [
  {"$match": {"$expr": {"$in": ["$nis", "$$daftar_nis"]}}}
  ],
  "as": "std"
}},
{"$unwind": "$std"},
{"$lookup": {
  "from": "c_jurusan",
  "let": {"kode_kelas": "$std.jurusan.kode_kelas", "tahun_ajaran": "$setting.kelas_x"},
  "pipeline": [
  {"$match": {
    "$and": [
    {"$expr": {"$eq": ["$kode_kelas", "$$kode_kelas"]}},
    {"$expr": {"$eq": ["$tahun_ajaran", "$$tahun_ajaran"]}}
    ]
  }}
  ],
  "as": "jurusan"
}},
{"$unwind": "$jurusan"},
{"$group": {
  "_id": "$kode_ekstra",
  "data": {"$push": {
    "nis": "$std.nis",
    "nama": "$std.nama",
    "kelas": {"$concat": [
    {"$switch": {
      "branches": [
      {
        "case": {"$eq": ["$std.tahun_masuk", "$setting.kelas_x"]}, 
        "then": "X"
      },
      {
        "case": {"$eq": ["$std.tahun_masuk", "$setting.kelas_xi"]}, 
        "then": "XI"
      },
      {
        "case": {"$eq": ["$std.tahun_masuk", "$setting.kelas_xii"]}, 
        "then": "XII"
      }
      ],
      "default": "Other"
    }}, " ", "$jurusan.nama_kelas"
    ]}
  }}
}}
]
`
if err := json.Unmarshal([]byte(plan), &pipe); err != nil {
  return bson.M{}, err
}
fmt.Printf("%#v\n", pipe)
cursor, err := db.StudentsEkstra.Aggregate(context.Background(), pipe)
if err != nil {
  return bson.M{}, err
}
for cursor.Next(context.Background()) {
  var elem bson.D
  if err := cursor.Decode(&elem); err != nil {
    continue
  }
  mapped := elem.Map()
  result = bson.M{
    mapped["_id"].(string): mapped["data"],
  }
  break
}
return result, nil
}

func (e *EkstraStruct) GetElevenStudents() ([]Siswa, error) {
  var result []Siswa
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return []Siswa{}, err
  }
  std, err := NewStudent().FindWithFilter(bson.M{"tahun_masuk": appSetting.TahunAjaran - 1})
  if err != nil {
    return []Siswa{}, err
  }
  for _, d := range std {
    if e.findElevenStudentsAndExist(bson.M{"nis": bson.M{"$in": bson.A{d.NIS}}, "tahun_ajaran": appSetting.TahunAjaran}) {
      continue
    }
    ElevenStudent := Siswa{
      Value: d.NIS,
      Label: strings.Join([]string{"XI", d.Jurusan.KodeKelas, d.Nama}, " "),
    }
    result = append(result, ElevenStudent)
  }
  return result, nil
}

func (e *EkstraStruct) findAndExist(filter bson.M) bool {
  var result EkstraStruct
  if err := db.Ekstra.FindOne(context.Background(), filter).Decode(&result); err != nil {
    return true
  }
  return result == (EkstraStruct{})
}

func (e *EkstraStruct) findElevenStudentsAndExist(filter bson.M) bool {
  cursor, err := db.StudentsEkstra.Find(context.Background(), filter)
  if err != nil {
    fmt.Println(err.Error())
    return true
  }
  return cursor.Next(context.Background())
}

func (e *EkstraStruct) marshal(data bson.M) (EkstraStruct, error) {
  var result EkstraStruct
  dByte, err := json.Marshal(data)
  if err != nil {
    return result, err
  }
  if err := json.Unmarshal(dByte, &result); err != nil {
    return result, err
  }
  return result, nil
}

func (e *EkstraStruct) InsertEkstraData(data []EkstraStruct) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  for _, d := range data {
    if e.findAndExist(bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_ekstra": d.KodeEkstra}) {
      return errors.New("kode ekstra" + d.KodeEkstra + "telah tersedia")
    }
    d.ID = primitive.NewObjectID()
    d.TahunAjaran = appSetting.TahunAjaran
    if _, err := db.Ekstra.InsertOne(context.Background(), d); err != nil {
      return err
    }
  }
  return nil
}

func (e *EkstraStruct) InsertStudentEkstra(kodeEkstra string, data []interface{}) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  filter := bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_ekstra": kodeEkstra}
  cursor, err := db.StudentsEkstra.Find(context.Background(), filter)
  if err != nil {
    return err
  }
  if !cursor.Next(context.Background()) {
    insert := bson.M{"kode_ekstra": kodeEkstra, "tahun_ajaran": appSetting.TahunAjaran, "nis": data}
    if _, err := db.StudentsEkstra.InsertOne(context.Background(), insert); err != nil {
      return err
    } else {
      return nil
    }
  }
  for _, d := range data {
    if e.findElevenStudentsAndExist(bson.M{"nis": bson.M{"$in": bson.A{d}}}) {
      fmt.Printf("%#v\n", d)
      return errors.New("error: duplikasi nis")
    }
    update := bson.M{"$push": bson.M{"nis": d}}
    if _, err := db.StudentsEkstra.UpdateOne(context.Background(), filter, update); err != nil {
      return err
    }
  }
  return nil
}

func (e *EkstraStruct) DeleteStudentEkstra(filter string, data []interface{}) error {
  appSetting, err := NewSetting().GetAppSetting()
  if err != nil {
    return err
  }
  filt := bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_ekstra": filter}
  update := bson.M{"$pull": bson.M{"nis": bson.M{"$in": data}}}
  if _, err := db.StudentsEkstra.UpdateOne(context.Background(), filt, update); err != nil {
    return err
  }
  return nil
}

func (e *EkstraStruct) UpdateEkstraData(filter map[string]interface{}, update map[string]interface{}) error {
  ID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
  if err != nil {
    return err
  }
  updateVal, err := e.marshal(update)
  if err != nil {
    return err
  }
  if _, err := db.Ekstra.UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": updateVal}); err != nil {
    return err
  }
  return nil
}

func (e *EkstraStruct) DeleteEkstraData(filter []bson.M) error {
  if _, err := db.Ekstra.DeleteMany(context.Background(), bson.M{"$or": filter}); err != nil {
    return err
  }
  return nil
}
