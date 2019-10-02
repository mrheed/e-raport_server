package models

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
)

type TugasStruct struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NIS          int                `json:"nis" bson:"nis"`
	Mapel        string             `json:"mapel" bson:"mapel"`
	NamaTugas    string             `json:"nama_tugas" bson:"nama_tugas"`
	TanggalTugas time.Time          `json:"tanggal_tugas" bson:"tanggal_tugas"`
	NilaiTugas   int                `json:"nilai_tugas" bson:"nilai_tugas"`
	Semester     int                `json:"semester" bson:"semester"`
	TahunAjaran  int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

type Task struct {
	Jurusan    string
	TahunMasuk string
	Data       *[]TugasStruct
	Result     *primitive.A
	Pipe       *[]bson.M
	Writer     http.ResponseWriter
}

func (t *Task) InsertTaskScore() error {
	var emptyStruct []interface{}
	for i, s := range *t.Data {
		for i1 := 0; i1 < reflect.ValueOf(s).NumField(); i1++ {
			gm := tool.MapStruct(s)
			gn := tool.GSName(s, i1)
			if tool.IsEmpty(gm[gn]) {
				return errors.New("Mohon mengisi " + gn + " pada data nomor " + strconv.Itoa(i+1))
			}
		}
		emptyStruct = append(emptyStruct, s)
	}
	_, err := db.NilaiTugas.InsertMany(context.Background(), emptyStruct)
	if err != nil {
		return err
	}
	return nil
}

func (t *Task) GetTaskStudent() error {
	pipe2 := `[
		{"$match": {
			"tahun_masuk": ` + t.TahunMasuk + `,
			"jurusan.kode_kelas": "` + t.Jurusan + `"
		}},
		{"$project": {
			"_id": 0,
			"value": "$nis",
			"label": "$nama"
		}}
	]`
	tool.ProcessPipeAggregate("Empty", t.Pipe, t.Writer)
	tool.ProcessPipeMiddleware(t.Pipe, pipe2)
	ok := tool.ProcessDataAggregate(t.Result, t.Writer, db.Student, *t.Pipe)
	if !ok {
		return errors.New("tidak dapat memproses data")
	}
	return nil
}

func (t *Task) GetTaskSubject() error {
	pipe2 := `[
		{"$match": {"$or": [
				{"mapel_kelas.value": "` + t.Jurusan + `"},
				{"mapel_kelas.value": "semua kelas"}
			]}
		},
		{"$project": {
			"_id": 0,
			"value": "$kode_mapel",
			"label": "$nama_mapel"
		}}
	]`
	tool.ProcessPipeAggregate("Empty", t.Pipe, t.Writer)
	tool.ProcessPipeMiddleware(t.Pipe, pipe2)
	ok := tool.ProcessDataAggregate(t.Result, t.Writer, db.Mapel, *t.Pipe)
	if !ok {
		return errors.New("tidak dapat memproses data")
	}
	return nil
}
