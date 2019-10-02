package routehandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"

	db "github.com/syahidnurrohim/restapi/database"
	mod "github.com/syahidnurrohim/restapi/models"
	tool "github.com/syahidnurrohim/restapi/utils"
)

func GetScoreController(w http.ResponseWriter, r *http.Request) {
	var result primitive.A
	var pipe []bson.M
	tipe := mux.Vars(r)["type"]
	variant := tool.GQry("var", r)
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if tipe != "" {
		tahunMasuk := tool.GQry("tahun_masuk", r)
		jurusan := tool.GQry("jurusan", r)
		materi := tool.GQry("materi", r)
		mapel := tool.GQry("mapel", r)
		siswa := tool.GQry("siswa", r)
		switch state := tool.GQry("state", r); state {
		case "remedy":
			remed := &mod.RemedyScore{
				Materi:         materi,
				Mapel:          mapel,
				Jurusan:        jurusan,
				Pipe:           &pipe,
				Result:         &result,
				ResponseWriter: w,
			}
			switch variant {
			case "student":
				thn, _ := strconv.Atoi(tahunMasuk)
				remed.TahunMasuk = thn
				remed.GetRemedyStudent(tipe)
			case "subject":
				appendedNis := appendCommaToMap("duh.nis", r)
				remed.NIS = appendedNis
				remed.GetRemedySubject(tipe)
			case "material":
				appendedNis := appendCommaToMap("duh.nis", r)
				remed.NIS = appendedNis
				remed.Mapel = tool.GQry("mapel", r)
				remed.GetRemedyMaterial()
			case "result":
				appendedNis := appendCommaToMap("nis", r)
				remed.NIS = appendedNis
				remed.GetRemedyScore(tipe)
			default:
				throw.Error(variant + " on parameter var isn't recognized by the server")
			}
		case "exam":
			exam := &mod.Exam{
				Materi: materi,
				Tipe:   tipe,
				Pipe:   &pipe,
				Result: &result,
			}
			switch variant {
			case "student":
				exam.Collection = db.Student
				data, err := exam.GetExamStudent(tahunMasuk, jurusan)
				if err != nil {
					throw.Error(err.Error())
					return
				}
				throw.Response(&data)
			case "subject":
				exam.Collection = db.Mapel
				data, err := exam.GetExamSubject(jurusan)
				if err != nil {
					throw.Error(err.Error())
					return
				}
				throw.Response(&data)
			case "material":
				data, err := exam.GetExamMaterial(mapel)
				if err != nil {
					throw.Error(err.Error())
					return
				}
				throw.Response(&data)
			case "result":
				exam.Collection = db.Student
				if err := exam.GetExamData(siswa, mapel, jurusan, tahunMasuk, w); err != nil {
					throw.Error(err.Error())
					return
				}
				throw.Response(&result)
			default:
				throw.Error(variant + " on parameter var isn't recognized by the server")
			}
		case "task":
			task := &mod.Task{
				Jurusan:    jurusan,
				TahunMasuk: tahunMasuk,
				Result:     &result,
				Pipe:       &pipe,
				Writer:     w,
			}
			switch variant {
			case "student":
				if err := task.GetTaskStudent(); err != nil {
					throw.Error(err.Error())
					return
				}
				throw.Response(&result)
			case "subject":
				if err := task.GetTaskSubject(); err != nil {
					throw.Error(err.Error())
					return
				}
				throw.Response(&result)
			// case "result":
			// 	if err := exam.GetExamData(); err != nil {
			// 		throw.Error(err.Error())
			// 		return
			// 	}
			// 	throw.Response(&result)
			default:
				throw.Error(variant + " on parameter var is unknown by the server")
			}
		default:
			if tool.GQry("type", r) != "" {
				exam := &mod.Exam{
					NIS: siswa,
				}
				switch tool.GQry("type", r) {
				case "per_mapel":
					exam.GetExamScoreWithAggregate(&result, pipe, w, "NilaiAggregate")
				case "total_mapel":
					exam.GetExamScoreWithAggregate(&result, pipe, w, "TotalNilai")
				default:
					json.NewEncoder(w).Encode("Invalid parameters combination")
				}
			} else {
				throw.Error(state + " on parameter state is unknown by the server")
			}
		}

	}

}

func UpdateScoreController(w http.ResponseWriter, r *http.Request) {
	var Data []mod.Remidi
	tipe := mux.Vars(r)["type"]
	remed := &mod.RemedyScore{}
	remed.ResponseWriter = w
	err := json.NewDecoder(r.Body).Decode(&Data)
	if err != nil {
		http.Error(w, tool.JSONErr("Error saat memproses data"), http.StatusMovedPermanently)
		return
	}
	err = remed.UpdateRemedyData(tipe, Data)
	if err != nil {
		http.Error(w, tool.JSONErr(err.Error()), http.StatusMovedPermanently)
		return
	}
	json.NewEncoder(w).Encode(tool.JSONGreen("Data berhasil diubah"))
}

func InsertScoreController(w http.ResponseWriter, r *http.Request) {
	tipe := mux.Vars(r)["type"]
	thrower := mod.NewThrower(w)
	thrower.StatusCode = http.StatusMovedPermanently
	switch tipe {
	case "UH", "PAS", "PTS":
		var Data []mod.Nilai
		exam := &mod.Exam{
			Tipe: tipe,
			Data: Data,
		}
		if err := json.NewDecoder(r.Body).Decode(&Data); err != nil {
			thrower.Error("tidak dapat memproses data")
			return
		}
		exam.Data = Data
		if exam.Tipe == "UH" {
			exam.Collection = db.NilaiUH
		} else {
			exam.Collection = db.NilaiPTPAS
		}
		if err := exam.InsertExamScore(); err != nil {
			thrower.Error(err.Error())
			return
		}
		thrower.Response(tool.JSONGreen("Data telah ditambahkan"))
	case "tugas":
		var Data []mod.TugasStruct
		task := &mod.Task{
			Data: &Data,
		}
		if err := json.NewDecoder(r.Body).Decode(&Data); err != nil {
			thrower.Error("tidak dapat memproses data")
			return
		}
		if err := task.InsertTaskScore(); err != nil {
			thrower.Error(err.Error())
			return
		}
		thrower.Response(tool.JSONGreen("Data telah ditambahkan"))
	}
}
