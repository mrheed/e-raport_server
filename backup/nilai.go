package routehandler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"

	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
)

type Nilai struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NIS            int                `json:"nis,omitempty" bson:"nis,omitempty"`
	Mapel          string             `json:"mapel,omitempty" bson:"mapel,omitempty"`
	Materi         string             `json:"materi,omitempty" bson:"materi,omitempty"`
	TanggalUlangan time.Time          `json:"tanggal_ulangan,omitempty" bson:"tanggal_ulangan,omitempty"`
	NilaiUlangan   int                `json:"nilai_ulangan,omitempty" bson:"nilai_ulangan,omitempty"`
	Remidi         bool               `json:"remidi" bson:"remidi"`
	TahunAjaran    int                `json:"tahun_ajaran,omitempty" bson:"tahun_ajaran,omitelmty"`
	Semester       int                `json:"semester,omitempty" bson:"semester,omitempty"`
	TelahRemidi    bool               `json:"telah_remidi" bson:"telah_remidi"`
	NilaiRemidi    int                `json:"nilai_remidi,omitempty" bson:"nilai_remidi,omitempty"`
	KKM            int                `json:"kkm,omitempty" bson:"kkm,omitempty"`
}

type NilaiPTPAS struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NIS            int                `json:"nis" bson:"nis"`
	TahunAjaran    int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
	Tipe           string             `json:"tipe" bson:"tipe"`
	Nilai          int                `json:"nilai_ulangan" bson:"nilai_ulangan"`
	Mapel          string             `json:"mapel" bson:"mapel"`
	KKM            int                `json:"kkm" bson:"kkm"`
	TanggalUlangan time.Time          `json:"tanggal_ulangan" bson:"tanggal_ulangan"`
	Remidi         bool               `json:"remidi" bson:"remidi"`
	TelahRemidi    bool               `json:"telah_remidi,omitempty" bson:"telah_remidi,omitempty"`
	NilaiRemidi    int                `json:"nilai_remidi,omitempty" bson:"nilai_remidi,omitempty"`
}

type Materi struct {
	KodeMateri string `json:"kode_materi,omitempty" bson:"kode_materi,omitempty"`
	NamaMateri string `json:"nama_materi,omitempty" bson:"nama_materi,omitempty"`
}

type Remidi struct {
	ID             int       `json:"_id,omitempty" bson:"_id,omitempty"`
	NilaiRemidi    int       `json:"nilai_remidi,omitempty" bson:"nilai_remidi,omitempty"`
	Materi         string    `json:"materi,omitempty" bson:"materi,omitempty`
	TanggalUlangan time.Time `json:"tanggal_ulangan,omitempty" bson:"tanggal_ulangan,omitempty"`
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

func AppendPTPASScoreController(w http.ResponseWriter, r *http.Request) {
	var PTPAS NilaiPTPAS
	err := json.NewDecoder(r.Body).Decode(&PTPAS)
	if err != nil {
		http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
		return
	}
	if tool.EmptyExecuted(PTPAS, w) {
		return
	}
	_, err = db.NilaiPTPAS.UpdateOne(context.Background(), bson.M{"nis": PTPAS.NIS}, bson.M{
		"$set": PTPAS,
	})
	if err != nil {
		http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
		return
	}
	json.NewEncoder(w).Encode(tool.JSONGreen("Data berhasil diubah"))
}

func InsertPTPASScoreController(w http.ResponseWriter, r *http.Request) {
	var PTPAS []NilaiPTPAS
	var SPTPAS NilaiPTPAS
	var newData []interface{}
	err := json.NewDecoder(r.Body).Decode(&PTPAS)
	if err != nil {
		http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
		return
	}
	for i1, s := range PTPAS {
		for i := 0; i < reflect.ValueOf(s).NumField(); i++ {
			if tool.IsEmpty(tool.MapStruct(s)[tool.GSName(s, i)]) {
				http.Error(w, tool.JSONErr("Mohon mengisi "+tool.GSName(s, i)+" pada data nomor "+strconv.Itoa(i1+1)), http.StatusMovedPermanently)
				return
			}
		}
		err := db.NilaiPTPAS.FindOne(context.Background(), bson.M{
			"nis":          s.NIS,
			"mapel":        s.Mapel,
			"tipe":         s.Tipe,
			"tahun_ajaran": s.TahunAjaran,
		}).Decode(&SPTPAS)
		if err == nil {
			http.Error(w, tool.JSONErr("Data dengan NIS "+strconv.Itoa(s.NIS)+" sudah tersedia di database"), http.StatusMovedPermanently)
			return
		}
		newData = append(newData, s)
	}

	_, err = db.NilaiPTPAS.InsertMany(context.Background(), newData)
	if err != nil {
		http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
		return
	}
	json.NewEncoder(w).Encode(tool.JSONGreen("Berhasil menambahkan data"))
}

func GetScoreController(w http.ResponseWriter, r *http.Request) {
	var result primitive.A
	var pipe []bson.M
	tipe := mux.Vars(r)["type"]
	variant := gQry("var", r)
	log.Println(tipe, variant)
	if gQry("state", r) == "remedy" {
		remed := &RemedyScore{}
		remed.Pipe = &pipe
		remed.Result = &result
		remed.ResponseWriter = w
		if tipe == "UH" {
			switch variant {
			case "student":
				thn, _ := strconv.Atoi(gQry("tahun_masuk", r))
				remed.Jurusan = gQry("jurusan", r)
				remed.TahunMasuk = thn
				remed.GetRemedyStudent()
			case "subject":
				appendedNis := appendCommaToMap("duh.nis", r)
				remed.NIS = appendedNis
				remed.GetRemedySubject()
			case "material":
				appendedNis := appendCommaToMap("duh.nis", r)
				remed.NIS = appendedNis
				remed.Mapel = gQry("mapel", r)
				remed.GetRemedyMaterial()
			case "result":
				appendedNis := appendCommaToMap("nis", r)
				remed.NIS = appendedNis
				remed.Materi = gQry("materi", r)
				remed.Mapel = gQry("mapel", r)
				remed.GetRemedyScore()
			}
		}
	}
	// db.NilaiPTPAS.Find(context.Background(), bson.M{"tipe": tipe})
}

func (r *RemedyScore) GetRemedyStudent() {
	pipe2 := `[{"$match": {"$and": 
		[{"jurusan.value": "` + r.Jurusan + `"}, 
		{"tahun_masuk": ` + strconv.Itoa(r.TahunMasuk) + `}]
	}}]`
	processPipeAggregate(agURI("StdRemedy"), r.Pipe, r.ResponseWriter)
	processPipeMiddleware(r.Pipe, pipe2)
	processDataAggregate(r.Result, r.ResponseWriter, db.Student, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(&r.Result)
}

func (r *RemedyScore) GetRemedySubject() {
	pipe2 := `[
		{"$lookup": {
			"from": "c_daftar_nilai_uh",
			"localField": "kode_mapel",
			"foreignField": "mapel",
			"as": "duh"
		}},
		{"$unwind": "$duh"},
		{"$match": {"$or": ` + string(r.NIS) + `}}
	]`
	processPipeAggregate(agURI("SubjectRemedy"), r.Pipe, r.ResponseWriter)
	processPipeMiddleware(r.Pipe, pipe2)
	processDataAggregate(r.Result, r.ResponseWriter, db.Mapel, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(&r.Result)
}

func (r *RemedyScore) GetRemedyMaterial() {
	pipe2 := `[
		{"$lookup": {
			"from": "c_daftar_nilai_uh",
			"localField": "kode_materi",
			"foreignField": "materi",
			"as": "duh"
		}},
		{"$unwind": "$duh"},
		{"$match": {"$and": [
			{"duh.remidi": true},
			{"duh.telah_remidi": false}, 
			{"duh.mapel": "` + r.Mapel + `"}, 
			{"$or": ` + string(r.NIS) + `}
		]}}
	]`
	processPipeAggregate(agURI("MaterialRemedy"), r.Pipe, r.ResponseWriter)
	processPipeMiddleware(r.Pipe, pipe2)
	processDataAggregate(r.Result, r.ResponseWriter, db.Kompetensi, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(r.Result)
}

func (r *RemedyScore) GetRemedyScore() {
	pipe2 := `[
		{"$match": {"$and": [
			{"mapel": "` + r.Mapel + `"}, 
			{"materi": "` + r.Materi + `"}, 
			{"$or": ` + string(r.NIS) + `}, 
			{"telah_remidi": false},
			{"remidi": true}]}},
		{"$lookup": {
			"from": "c_students",
			"localField": "nis",
			"foreignField": "nis",
			"as": "std"
		}},
		{"$unwind": "$std"}
	]`
	processPipeAggregate(agURI("ScoreRemedy"), r.Pipe, r.ResponseWriter)
	processPipeMiddleware(r.Pipe, pipe2)
	processDataAggregate(r.Result, r.ResponseWriter, db.NilaiUH, *r.Pipe)
	json.NewEncoder(r.ResponseWriter).Encode(&r.Result)
}

func AppendRemedyScoreController(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		if _, err := tool.VerifyHeader(AllUser, r, w); err == nil {
			var remidiData []Remidi
			var errData []string
			json.NewDecoder(r.Body).Decode(&remidiData)
			for i1, rData := range remidiData {
				var nData Nilai
				for i := 0; i < reflect.ValueOf(nData).NumField(); i++ {
					if tool.IsEmpty(nData) {
						http.Error(w, tool.JSONErr("Mohon mengisi "+tool.GSName(nData, i)+" pada data nomor "+strconv.Itoa(i1+1)), http.StatusMovedPermanently)
						return
					}
				}
				err := db.NilaiUH.FindOneAndUpdate(context.Background(),
					bson.M{"nis": rData.ID, "materi": rData.Materi, "tanggal_ulangan": rData.TanggalUlangan, "remidi": true, "telah_remidi": false},
					bson.M{"$set": bson.M{"nilai_remidi": rData.NilaiRemidi, "telah_remidi": true}},
				).Decode(&nData)
				if err != nil {
					errData = append(errData, strconv.Itoa(rData.ID))
				}
			}
			if len(errData) == 0 {
				json.NewEncoder(w).Encode(tool.JSONGreen("Data berhasil di update"))
			} else if len(errData) == len(remidiData) {
				http.Error(w, tool.JSONErr("Data tidak diproses"), http.StatusMovedPermanently)
			} else {
				json.NewEncoder(w).Encode(tool.JSONGreen(strconv.Itoa(len(remidiData)-len(errData)) + "data telah diupdate" + strconv.Itoa(len(errData)) + "diabaikan dengan nis : " + strings.Join(errData, ", ")))
			}
		}
	}
}

func AppendExamScoreController(w http.ResponseWriter, r *http.Request) {
	if _, err := tool.VerifyHeader(AllUser, r, w); err == nil {
		var roughData []Nilai
		var newData []interface{}
		var existData []string
		json.NewDecoder(r.Body).Decode(&roughData)
		for i1, data := range roughData {
			var Dec Nilai
			for i := 0; i < reflect.ValueOf(data).NumField(); i++ {
				if tool.IsEmpty(data) {
					http.Error(w, tool.JSONErr("Mohon mengisi "+tool.GSName(data, i)+" pada data nomor "+strconv.Itoa(i1+1)), http.StatusMovedPermanently)
					return
				}
			}
			if data.NilaiUlangan < data.KKM {
				data.Remidi = true
				data.TelahRemidi = false
			}
			db.NilaiUH.FindOne(context.Background(), bson.M{"$and": []interface{}{bson.M{"nis": data.NIS}, bson.M{"materi": data.Materi}}}).Decode(&Dec)
			if Dec != (Nilai{}) {
				existData = append(existData, strconv.Itoa(Dec.NIS))
			} else {
				newData = append(newData, data)
			}
		}
		res, err := db.NilaiUH.InsertMany(context.Background(), newData)
		if err != nil {
			http.Error(w, tool.JSONErr("Data tidak ditambahkan karena sudah tersedia di database"), http.StatusInternalServerError)
			return
		}
		if len(existData) != 0 {
			json.NewEncoder(w).Encode(tool.JSONGreen(strconv.Itoa(len(res.InsertedIDs)) + "data ditambahkan dan" + strconv.Itoa(len(existData)) + "diabaikan dengan NIS: " + strings.Join(existData, ", ")))
		}
	}
}

func GetExamScoreController(w http.ResponseWriter, r *http.Request) {
	var result primitive.A
	pipe := []bson.M{}

	if gQry("type", r) == "per_mapel" {
		processPipeAggregate(agURI("NilaiAggregate"), &pipe, w)
		pipe2 := `[
			{"$match": {"tahun_ajaran": ` + gQry("tahun_ajaran", r) + `}}
		]`
		processPipeMiddleware(&pipe, pipe2)
		processDataAggregate(&result, w, db.NilaiUH, pipe)
	} else if gQry("type", r) == "total_mapel" {
		processPipeAggregate(agURI("TotalNilai"), &pipe, w)
		pipe2 := `[
			{
				"$lookup": {
					"from": "c_daftar_nilai_uh",
					"localField": "nis",
					"foreignField": "nis",
					"as": "duh"
				}
			},
			{
				"$unwind": "$duh"
			},
			{
				"$match": {"duh.tahun_ajaran": ` + gQry("tahun_ajaran", r) + `}
			}
			]`
		processPipeMiddleware(&pipe, pipe2)
		processDataAggregate(&result, w, db.Student, pipe)
	} else if gQry("status", r) == "remedy" {

		if (gQry("jurusan", r) != "") && gQry("tahun_masuk", r) != "" {
			if gQry("siswa", r) != "" {
				mb := appendCommaToMap("duh.nis", r)
				mc := appendCommaToMap("nis", r)
				if gQry("mapel", r) != "" {
					if gQry("materi", r) != "" {
						if gQry("variety", r) == "read_remedy_score" {
							pipe2 := `[
								{"$match": {"$and": [
									{"mapel": "` + gQry("mapel", r) + `"}, 
									{"materi": "` + gQry("materi", r) + `"}, 
									{"$or": ` + string(mc) + `}, 
									{"telah_remidi": false},
									{"remidi": true}]}},
								{"$lookup": {
									"from": "c_students",
									"localField": "nis",
									"foreignField": "nis",
									"as": "std"
								}},
								{"$unwind": "$std"}
							]`
							processPipeAggregate(agURI("ScoreRemedy"), &pipe, w)
							processPipeMiddleware(&pipe, pipe2)
							processDataAggregate(&result, w, db.NilaiUH, pipe)
						} else {
							json.NewEncoder(w).Encode("Error: Invalid parameter combination")
							return
						}
					} else {
						// Materi scond condition
						if gQry("variety", r) == "read_remedy_material" {
							pipe2 := `[
								{"$lookup": {
									"from": "c_daftar_nilai_uh",
									"localField": "kode_materi",
									"foreignField": "materi",
									"as": "duh"
								}},
								{"$unwind": "$duh"},
								{"$match": {"$and": [
									{"duh.remidi": true},
									{"duh.telah_remidi": false}, 
									{"duh.mapel": "` + gQry("mapel", r) + `"}, 
									{"$or": ` + string(mb) + `}
								]}},
								{"$lookup": {
									"from": "c_students",
									"localField": "duh.nis",
									"foreignField": "nis",
									"as": "std"
								}},
								{"$unwind": "$std"},
								{"$match": {"std.jurusan.value": "` + gQry("jurusan", r) + `"}}
							]`
							processPipeAggregate(agURI("MaterialRemedy"), &pipe, w)
							processPipeMiddleware(&pipe, pipe2)
							processDataAggregate(&result, w, db.Kompetensi, pipe)
						} else {
							json.NewEncoder(w).Encode("Error: Invalid parameter combination")
							return
						}
					}

				} else {
					// Mapel second condition
					if gQry("variety", r) == "read_remedy_subject" {
						pipe2 := `[
							{"$lookup": {
								"from": "c_daftar_nilai_uh",
								"localField": "kode_mapel",
								"foreignField": "mapel",
								"as": "duh"
							}},
							{"$unwind": "$duh"},
							{"$match": {"$or": ` + string(mb) + `}}
						]`
						processPipeAggregate(agURI("SubjectRemedy"), &pipe, w)
						processPipeMiddleware(&pipe, pipe2)
						processDataAggregate(&result, w, db.Mapel, pipe)
					} else {
						json.NewEncoder(w).Encode("Error: Invalid parameter combination")
						return
					}
				}

			} else {
				// Siswa second condition
				if gQry("variety", r) == "read_remedy_std" {
					pipe2 := `[{"$match": {"$and": 
						[{"jurusan.value": "` + gQry("jurusan", r) + `"}, 
						{"tahun_masuk": ` + gQry("tahun_masuk", r) + `}]
					}}]`
					processPipeAggregate(agURI("StdRemedy"), &pipe, w)
					processPipeMiddleware(&pipe, pipe2)
					processDataAggregate(&result, w, db.Student, pipe)
				} else {
					json.NewEncoder(w).Encode("Error: Invalid parameter combination")
					return
				}
			}
		} else {
			json.NewEncoder(w).Encode("Error: Parameter jurusan tidak tercantum")
			return
		}

	} else {
		// Jurusan second condition
		json.NewEncoder(w).Encode("Invalid URL Parameter")
		return
	}
	// cursor.All(context.Background(), &result)
	json.NewEncoder(w).Encode(&result)
}

func appendCommaToMap(keyName string, r *http.Request) []byte {
	var nisSlice []map[string]int
	for _, s := range strings.Split(gQry("siswa", r), ",") {
		nis, _ := strconv.Atoi(s)
		dt := map[string]int{
			keyName: nis,
		}
		nisSlice = append(nisSlice, dt)
	}
	mb, _ := json.Marshal(nisSlice)
	return mb
}

func agURI(filename string) string {
	return "./database/json/" + filename + ".json"
}

func gQry(key string, r *http.Request) string {
	return r.URL.Query().Get(key)
}

func processDataAggregate(result *primitive.A, w http.ResponseWriter, collection *mongo.Collection, pipe []bson.M) {
	cursor, err := collection.Aggregate(context.Background(), pipe)
	if err != nil {
		http.Error(w, tool.JSONErr("Ups!, ada yang salah dengan servernya"), http.StatusInternalServerError)
		return
	}
	for cursor.Next(context.Background()) {
		var elem bson.D
		cursor.Decode(&elem)
		*result = append(*result, elem.Map())
	}
}

func processPipeMiddleware(pipe *[]bson.M, secondPipe string) {
	prependQ := []bson.M{}
	json.Unmarshal([]byte(secondPipe), &prependQ)
	*pipe = append(prependQ, *pipe...)
}

func processPipeAggregate(location string, pipe *[]bson.M, w http.ResponseWriter) {
	plan, _ := locateReadFile(location)
	json.Unmarshal(plan, &pipe)
}

func locateReadFile(location string) ([]byte, error) {
	path, err := filepath.Abs(location)
	if err != nil {
		return nil, err
	}
	plan, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return plan, nil
}
