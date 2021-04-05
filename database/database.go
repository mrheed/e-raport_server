package database

import (
  "context"
  "fmt"
  "reflect"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/bsontype"

  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

// DB information listed here
var DB = map[string]string{
  "dbname": "e_raport",
  "cu":     "c_users",
  "ce":     "c_ekstra",
  "cses":   "c_sessions",
  "cs":     "c_students",
  "ct":     "c_teacher",
  "cm":     "c_mapel",
  "ck":     "c_kelas",
  "cj":     "c_jurusan",
  "ckd":    "c_materi",
  "cse":    "c_students_ekstra",
  "cdu":    "c_daftar_nilai_uh",
  "ctgs":   "c_daftar_nilai_tugas",
  "cas":    "c_application_setting",
  "cptpas": "c_daftar_nilai_pts_pas",
}

// Database url
var local = "mongodb://localhost:27017"
var cloud = "mongodb+srv://rapor:erapor@cluster0-ousnv.mongodb.net/test?retryWrites=true&w=majority"
var clientURI = local

var (
  client *mongo.Client
  // Base db handler
  Base *mongo.Database
  // Ekstra mgo
  Ekstra *mongo.Collection
  // Session mgo
  Session *mongo.Collection
  // User mgo
  User *mongo.Collection
  // Student mgo
  Student *mongo.Collection
  // Teacher mgo
  Teacher *mongo.Collection
  // Mapel mgo
  Mapel *mongo.Collection
  // Kelas mgo
  Kelas *mongo.Collection
  // Jurusan mgo
  Jurusan *mongo.Collection
  // Kompetensi mgo
  Kompetensi *mongo.Collection
  // NilaiUH mgo
  NilaiUH *mongo.Collection
  // AppSetting mgo
  AppSetting *mongo.Collection
  // NilaiPTPAS mgo
  NilaiPTPAS *mongo.Collection
  // NilaiTugas mgo
  NilaiTugas *mongo.Collection
  // StudentsEkstra mgo
  StudentsEkstra *mongo.Collection
)

func Disconnect() {
  client.Disconnect(context.Background())
}

// InitDBAndCollection func
func InitDBAndCollection() {

  rb := bson.NewRegistryBuilder()
  rb.RegisterTypeMapEntry(bsontype.EmbeddedDocument, reflect.TypeOf(bson.M{}))
  client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(clientURI).SetRegistry(rb.Build()))
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  Base = client.Database(DB["dbname"])

  Session = Base.Collection(DB["cses"])
  User = Base.Collection(DB["cu"])
  Ekstra = Base.Collection(DB["ce"])
  Student = Base.Collection(DB["cs"])
  Teacher = Base.Collection(DB["ct"])
  Kelas = Base.Collection(DB["ck"])
  Jurusan = Base.Collection(DB["cj"])
  Mapel = Base.Collection(DB["cm"])
  Kompetensi = Base.Collection(DB["ckd"])
  NilaiUH = Base.Collection(DB["cdu"])
  AppSetting = Base.Collection(DB["cas"])
  NilaiPTPAS = Base.Collection(DB["cptpas"])
  NilaiTugas = Base.Collection(DB["ctgs"])
  StudentsEkstra = Base.Collection(DB["cse"])

  tool.CreateUniqueIndex("nis", Student)
  tool.CreateUniqueIndex("nip", Teacher)
  tool.CreateUniqueIndex("username", User)
  tool.CreateUniqueIndex("kode_kelas", Kelas)

}
