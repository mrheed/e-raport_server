package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Nilai struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NIS            int                `json:"nis,omitempty" bson:"nis,omitempty"`
	Mapel          string             `json:"mapel,omitempty" bson:"mapel,omitempty"`
	Materi         string             `json:"materi,omitempty" bson:"materi,omitempty"`
	TanggalUlangan time.Time          `json:"tanggal_ulangan,omitempty" bson:"tanggal_ulangan,omitempty"`
	NilaiUlangan   int                `json:"nilai_ulangan,omitempty" bson:"nilai_ulangan,omitempty"`
	Remidi         bool               `json:"remidi" bson:"remidi"`
	Tipe           string             `json:"tipe" bson:"tipe"`
	TahunAjaran    int                `json:"tahun_ajaran,omitempty" bson:"tahun_ajaran,omitelmty"`
	Semester       int                `json:"semester,omitempty" bson:"semester,omitempty"`
	TelahRemidi    bool               `json:"telah_remidi" bson:"telah_remidi"`
	NilaiRemidi    int                `json:"nilai_remidi,omitempty" bson:"nilai_remidi,omitempty"`
	KKM            int                `json:"kkm,omitempty" bson:"kkm,omitempty"`
}

type Exam struct {
	Collection *mongo.Collection
	Data       []Nilai
	Tipe       string
	NIS        string
	Jurusan    string
	Materi     string
	Pipe       *[]bson.M
	Result     *primitive.A
}

func (e *Exam) InsertExamScore() error {
	var emptyType []interface{}
	var existData []string
	var filter bson.M
	for i, s := range e.Data {
		var sData Nilai
		for i1 := 0; i1 < reflect.ValueOf(s).NumField(); i1++ {
			mp := tool.MapStruct(s)
			gn := tool.GSName(s, i1)
			if tool.IsEmpty(mp[gn]) {
				return errors.New("Mohon mengisi " + gn + " pada data nomor " + strconv.Itoa(i+1))
			}
		}
		if s.NilaiUlangan < s.KKM {
			s.Remidi = true
		}
		filter = bson.M{
			"nis":          s.NIS,
			"mapel":        s.Mapel,
			"materi":       s.Materi,
			"tahun_ajaran": s.TahunAjaran,
			"semester":     s.Semester,
		}
		if e.Tipe == "UH" {
			filter["materi"] = s.Materi
		} else {
			filter["tipe"] = s.Tipe
		}
		err := e.Collection.FindOne(context.Background(), filter).Decode(&sData)
		if err == nil {
			existData = append(existData, strconv.Itoa(sData.NIS))
		}
		emptyType = append(emptyType, s)
	}
	if len(existData) != 0 {
		return errors.New("Data dengan NIS : " + strings.Join(existData, ", ") + "telah terdaftar di basis data")
	}
	_, err := e.Collection.InsertMany(context.Background(), emptyType)
	if err != nil {
		return err
	}
	return nil
}

func (e *Exam) GetExamStudent(tahunMasuk string, jurusan string) ([]bson.M, error) {
	var result []bson.M
	tahunMasukInt, err := strconv.Atoi(tahunMasuk)
	if err != nil {
		return []bson.M{}, err
	}
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return []bson.M{}, err
	}
	std, err := NewStudent().FindWithFilter(bson.M{"tahun_masuk": tahunMasukInt, "jurusan.kode_kelas": jurusan, "jurusan.tahun_ajaran": appSetting.TahunAjaran})
	fmt.Printf("%+v\n", tahunMasuk, jurusan)
	if err != nil {
		return []bson.M{}, err
	}
	for _, d := range std {
		result = append(result, bson.M{"value": d.NIS, "label": d.Nama})
	}
	return result, nil
}

func (e *Exam) GetExamSubject(jurusan string) ([]bson.M, error) {
	var result []bson.M
	s := strings.Split(jurusan, " ")
	classOnYear, err := NewMisc().GetClassOnSchoolYear()
	if err != nil {
		return []bson.M{}, err
	}
	sbj, err := NewMapel().FindWithFilter(bson.M{
		"mapel_kelas": bson.M{"$elemMatch": bson.M{"jurusan": s[1], "tahun_ajaran": classOnYear.SchoolYearOnGrade[s[0]]}},
	})
	for _, d := range sbj {
		result = append(result, bson.M{"value": d.KodeMapel, "label": d.NamaMapel})
	}
	return result, nil
}

func (e *Exam) GetExamMaterial(mapel string) ([]bson.M, error) {
	var result []bson.M
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return []bson.M{}, err
	}
	materi, err := NewMateri().FindWithFilter(bson.M{"nama_mapel.kode_mapel": mapel, "nama_mapel.tahun_ajaran": appSetting.TahunAjaran, "tahun_ajaran": appSetting.TahunAjaran})
	if err != nil {
		return []bson.M{}, err
	}
	for _, d := range materi {
		result = append(result, bson.M{"label": d.NamaMateri, "value": d.KodeMateri})
	}
	return result, nil
}

func (e *Exam) GetExamData(NIS string, mapel string, jurusan string, tahun_masuk string, w http.ResponseWriter) error {
	var dbNilai *mongo.Collection
	var filter bson.M
	var AnotherNIS []string
	var MappedNIS []map[string]int
	splitted := strings.Split(NIS, ",")
	for _, s := range splitted {
		NISVal, _ := strconv.Atoi(s)
		MappedNIS = append(MappedNIS, map[string]int{"nis": NISVal})
	}
	if e.Tipe == "UH" {
		dbNilai = db.NilaiUH
		filter = bson.M{"$or": MappedNIS, "mapel": mapel, "materi": e.Materi}
	} else {
		dbNilai = db.NilaiPTPAS
		filter = bson.M{"$or": MappedNIS, "mapel": mapel, "tipe": e.Tipe}
	}
	cursor, err := dbNilai.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	for cursor.Next(context.Background()) {
		var Nilai Nilai
		err := cursor.Decode(&Nilai)
		if err != nil {
			continue
		}
		AnotherNIS = append(AnotherNIS, strconv.Itoa(Nilai.NIS))
	}
	plan, err := json.Marshal(&MappedNIS)
	if err != nil {
		return err
	}
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return err
	}
	pipe2 := `[
  {"$match": {"$and": [
  {"$or": ` + string(plan) + `},
  {"nis": {"$nin": [` + strings.Join(AnotherNIS, ",") + `]}},
  {"jurusan.kode_kelas": "` + jurusan + `"},
  {"jurusan.tahun_ajaran": ` + strconv.Itoa(appSetting.TahunAjaran) + `},
  {"tahun_masuk": ` + tahun_masuk + `}
  ]}},
  {"$project": {"_id": 0, "nis": 1, "nama": 1}}
  ]`
	tool.ProcessPipeAggregate("Empty", e.Pipe, w)
	tool.ProcessPipeMiddleware(e.Pipe, pipe2)
	tool.ProcessDataAggregate(e.Result, w, e.Collection, *e.Pipe)
	fmt.Printf("%+v\n", pipe2)
	return nil
}

func (e *Exam) GetExamScoreWithAggregate(result *primitive.A, pipe []bson.M, w http.ResponseWriter, filename string) {
	var pipe2 string
	tool.ProcessPipeAggregate(tool.AgURI(filename), &pipe, w)
	if filename == "NilaiAggregate" {
		var arrayNIS []string
		nis := strings.Split(e.NIS, ",")
		for _, d := range nis {
			arrayNIS = append(arrayNIS, `{"nis":`+d+`}`)
		}
		strNIS := strings.Join(arrayNIS, ",")
		pipe2 = `[
    {"$match": {"$or": [
    ` + strNIS + `
    ]}}
    ]`
	} else if filename == "TotalNilai" {
		pipe2 = `[
    ]`
	}
	tool.ProcessPipeMiddleware(&pipe, pipe2)
	tool.ProcessDataAggregate(result, w, db.Student, pipe)
	json.NewEncoder(w).Encode(&result)
}
