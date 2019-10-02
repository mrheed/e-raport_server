package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Remidi struct {
	ID            int       `json:"_id,omitempty" bson:"_id,omitempty"`
	NilaiRemidi   int       `json:"nilai_remidi" bson:"nilai_remidi"`
	Materi        string    `json:"materi" bson:"materi"`
	Mapel         string    `json:"mapel" bson:"mapel"`
	TanggalRemidi time.Time `json:"tanggal_remidi" bson:"tanggal_remidi"`
}

type RemedyScore struct {
	NIS            []byte
	Jurusan        string
	Mapel          string
	Materi         string
	TahunMasuk     int
	Result         *primitive.A
	Pipe           *[]bson.M
	ResponseWriter http.ResponseWriter
}

func (r *RemedyScore) GetRemedyStudent(tipe string) {
	var collName string
	pipe3 := ``
	if tipe == "UH" {
		collName = `c_daftar_nilai_uh`
	} else {
		collName = `c_daftar_nilai_pts_pas`
		pipe3 = `,{"$match": {"duh.tipe": "` + tipe + `"}}`
	}
	pipe2 := `[
		{"$lookup": {
			"from": "c_application_setting",
			"pipeline": [],
			"as": "setting"
		}},
		{"$unwind": "$setting"},
		{"$match": 
			{"$and": [
				{"jurusan.kode_kelas": "` + r.Jurusan + `"}, 
				{"$expr": {"$eq": ["$jurusan.tahun_ajaran", "$setting.tahun_ajaran"]}},
				{"tahun_masuk": ` + strconv.Itoa(r.TahunMasuk) + `}
			]}
		},
		{"$lookup": {
			"from": "` + collName + `",
			"localField": "nis",
			"foreignField": "nis",
			"as": "duh"
		}},
		{"$unwind": "$duh"}` + pipe3 + `
	]`
	tool.ProcessPipeAggregate(tool.AgURI("StdRemedy"), r.Pipe, r.ResponseWriter)
	tool.ProcessPipeMiddleware(r.Pipe, pipe2)
	if !tool.ProcessDataAggregate(r.Result, r.ResponseWriter, db.Student, *r.Pipe) {
		return
	}
	fmt.Printf("%+v\n", *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(&r.Result)
}

func (r *RemedyScore) GetRemedySubject(tipe string) {
	var collName string
	pipe3 := ``
	if tipe == "UH" {
		collName = `c_daftar_nilai_uh`
	} else {
		collName = `c_daftar_nilai_pts_pas`
		pipe3 = `,{"$match": {"duh.tipe": "` + tipe + `"}}`
	}
	pipe2 := `[
		{"$lookup": {
			"from": "c_application_setting",
			"pipeline": [],
			"as": "setting"
		}},
		{"$unwind": "$setting"},
		{"$match": {"$expr": {"$eq": ["$tahun_ajaran", "$setting.tahun_ajaran"]}}},
		{"$lookup": {
			"from": "` + collName + `",
			"let": {"tahun_ajaran": "$setting.tahun_ajaran", "kode_mapel": "$kode_mapel"},
			"pipeline": [
				{"$match": {"$and": [{"$expr": {"$eq": ["$tahun_ajaran", "$$tahun_ajaran"]}}, {"$expr": {"$eq": ["$mapel", "$$kode_mapel"]}}] }}
			],
			"as": "duh"
		}},
		{"$unwind": "$duh"},
		{"$match": {"$or": ` + string(r.NIS) + `}}` + pipe3 + `
	]`
	tool.ProcessPipeAggregate(tool.AgURI("SubjectRemedy"), r.Pipe, r.ResponseWriter)
	tool.ProcessPipeMiddleware(r.Pipe, pipe2)
	tool.ProcessDataAggregate(r.Result, r.ResponseWriter, db.Mapel, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(&r.Result)
}

func (r *RemedyScore) GetRemedyMaterial() {
	pipe2 := `[
		{"$lookup": {
			"from": "c_application_setting",
			"pipeline": [],
			"as": "setting"
		}},
		{"$unwind": "$setting"},
		{"$match": {"$expr": {"$eq": ["$tahun_ajaran", "$setting.tahun_ajaran"]}}},
		{"$lookup": {
			"from": "c_daftar_nilai_uh",
			"let": {"tahun_ajaran": "$setting.tahun_ajaran", "kode_materi": "$kode_materi"},
			"pipeline": [
				{"$match": {"$and": [
						{"$expr": {"$eq": ["$tahun_ajaran", "$$tahun_ajaran"]}}, 
						{"$expr": {"$eq": ["$materi", "$$kode_materi"]}}
				]}}
			],
			"as": "duh"
		}},
		{"$unwind": "$duh"},
		{"$match": {"$or": ` + string(r.NIS) + `}}
	]`
	tool.ProcessPipeAggregate(tool.AgURI("MaterialRemedy"), r.Pipe, r.ResponseWriter)
	tool.ProcessPipeMiddleware(r.Pipe, pipe2)
	tool.ProcessDataAggregate(r.Result, r.ResponseWriter, db.Kompetensi, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(r.Result)
}

func (r *RemedyScore) GetRemedyScore(tipe string) {
	var coll *mongo.Collection
	pipe3 := ``
	if tipe == "UH" {
		coll = db.NilaiUH
		pipe3 = `{"materi": "` + r.Materi + `"},`
	} else {
		coll = db.NilaiPTPAS
		pipe3 = `{"tipe": "` + tipe + `"},`
	}
	pipe2 := `[
		{"$lookup": {
			"from": "c_application_setting",
			"pipeline": [],
			"as": "setting"
		}},
		{"$unwind": "$setting"},
		{"$match": {"$and": [
			{"mapel": "` + r.Mapel + `"},  
			` + pipe3 + `
			{"$or": ` + string(r.NIS) + `}, 
			{"telah_remidi": false},
			{"$expr": {"$eq": ["$tahun_ajaran", "$setting.tahun_ajaran"]}},
			{"remidi": true}]
		}},
		{"$lookup": {
			"from": "c_students",
			"localField": "nis",
			"foreignField": "nis",
			"as": "std"
		}},
		{"$unwind": "$std"}
	]`
	tool.ProcessPipeAggregate(tool.AgURI("ScoreRemedy"), r.Pipe, r.ResponseWriter)
	tool.ProcessPipeMiddleware(r.Pipe, pipe2)
	tool.ProcessDataAggregate(r.Result, r.ResponseWriter, coll, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(&r.Result)
}

func (r *RemedyScore) UpdateRemedyData(tipe string, data []Remidi) error {
	for i, s := range data {
		var N Nilai
		var filter bson.M
		coll := IEReturnMongo(tipe == "UH", db.NilaiUH, db.NilaiPTPAS)
		for i1 := 0; i1 < reflect.ValueOf(s).NumField(); i1++ {
			mp := tool.MapStruct(s)
			gn := tool.GSName(s, i1)
			if tool.IsEmpty(mp[gn]) {
				return errors.New("Mohon mengisi " + gn + " pada data nomor " + strconv.Itoa(i+1))
			}
		}
		bytes := []byte(`{
			` + IEReturnString(tipe == "UH", `"materi": "`+s.Materi+`",`, ``) + `
			"nis": ` + strconv.Itoa(s.ID) + `,
			"mapel": "` + s.Mapel + `",
			"remidi": true,
			"telah_remidi": false
		}`)
		err := json.Unmarshal(bytes, &filter)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		err = coll.FindOne(context.Background(), bson.M{
			"nis":          s.ID,
			"mapel":        s.Mapel,
			"tipe":         tipe,
			"remidi":       true,
			"telah_remidi": true,
		}).Decode(&N)
		if err == nil {
			return errors.New("data sudah tersedia di basis data")
		}
		_, err = coll.UpdateOne(context.Background(), filter, bson.M{
			"$set": bson.M{
				"tanggal_remidi": s.TanggalRemidi,
				"telah_remidi":   true,
				"nilai_remidi":   s.NilaiRemidi,
			}})
		if err != nil {
			return err
		}
	}
	return nil
}

func IEReturnString(condition bool, result1 string, result2 string) string {
	if condition {
		return result1
	}
	return result2
}

func IEReturnMongo(condition bool, result1 *mongo.Collection, result2 *mongo.Collection) *mongo.Collection {
	if condition {
		return result1
	}
	return result2
}
