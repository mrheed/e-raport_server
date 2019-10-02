package routehandler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	db "github.com/syahidnurrohim/restapi/database"
	mod "github.com/syahidnurrohim/restapi/models"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Student struct
type Student struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NIS          int                `json:"nis,omitempty" bson:"nis,omitsempty"`
	Nama         string             `json:"nama,omitempty" bson:"nama,omitempty"`
	Jurusan      tool.CustomSelect  `json:"jurusan,omitempty" bson:"jurusan,omitempty"`
	JenisKelamin string             `json:"jeniskelamin,omitempty" bson:"jeniskelamin,omitempty"`
	TahunMasuk   int                `json:"tahun_masuk,omitempty" bson:"tahun_masuk,omitempty"`
}

// UpdateStudent struct
type UpdateStudent struct {
	Update Student     `json:"update" bson:"update"`
	Filter tool.Filter `json:"filter" bson:"filter"`
}

// GetStudentsController func
func GetStudentsController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(AllUser, r, w); err != nil {
	}
	student := mod.NewStudent()
	appSetting, err := mod.NewSetting().GetAppSetting()
	if err != nil {
		throw.Error(err.Error())
		return
	}
	if jurusan := gQry("jurusan", r); jurusan != "" {
		if tahun_masuk := gQry("tahun_masuk", r); tahun_masuk != "" {
			var throwData []bson.M
			tahunMasuk, err := strconv.Atoi(tahun_masuk)
			if err != nil {
				throw.Error(err.Error())
				return
			}
			data, err := student.FindWithFilter(bson.M{"jurusan.kode_kelas": jurusan, "tahun_masuk": tahunMasuk})
			if err != nil {
				throw.Error(err.Error())
				return
			}
			jurusan, err := mod.NewVocation().GetSingleVocation(bson.M{"kode_kelas": jurusan, "tahun_ajaran": appSetting.TahunAjaran})
			if err != nil {
				throw.Error(err.Error())
				return
			}
			for _, d := range data {
				tmpData := bson.M{
					"_id":          d.ID,
					"nis":          d.NIS,
					"nama":         d.Nama,
					"jurusan":      bson.M{"value": jurusan.KodeKelas, "label": jurusan.NamaKelas},
					"jeniskelamin": d.JenisKelamin,
					"tahun_masuk":  d.TahunMasuk,
				}
				throwData = append(throwData, tmpData)
			}
			throw.Response(&throwData)
			return
		}
		data, err := student.FindWithFilter(bson.M{"jurusan.kode_kelas": jurusan, "jurusan.tahun_ajaran": appSetting.TahunAjaran})
		if err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(&data)
		return
	}
	data, err := student.GetAllStudents()
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(&data)
}

// InsertStudentController func
func InsertStudentController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
	}
	student := mod.NewStudent()
	inserted, err := student.InsertStudents(r)
	if err != nil {
		throw.Error(err.Error())
		return
	}
	throw.Response(tool.JSONGreen(strconv.Itoa(inserted) + "data telah ditambahkan"))
}

// UpdateDeleteStudentController func
func UpdateDeleteStudentController(w http.ResponseWriter, r *http.Request) {
	throw := mod.NewThrower(w)
	throw.StatusCode = http.StatusMovedPermanently
	if _, err := tool.VerifyHeader(AllUser, r, w); err != nil {
	}
	student := mod.NewStudent()
	if r.Method == "DELETE" {
		var std []primitive.M
		json.NewDecoder(r.Body).Decode(&std)
		cursor, err := db.Student.DeleteMany(context.Background(), bson.D{primitive.E{Key: "$or", Value: std}})
		if err != nil {
			http.Error(w, tool.JSONErr(err.Error()), http.StatusInternalServerError)
			return
		}
		deletedCount := strconv.FormatInt(cursor.DeletedCount, 10)
		json.NewEncoder(w).Encode(tool.JSONGreen(deletedCount + " records has been deleted"))
	}

	if r.Method == "PUT" {
		var emptyStruct map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&emptyStruct); err != nil {
			throw.Error(err.Error())
			return
		}
		if err := student.UpdateStudent(emptyStruct["filter"].(map[string]interface{}), emptyStruct["update"].(map[string]interface{})); err != nil {
			throw.Error(err.Error())
			return
		}
		throw.Response(tool.JSONGreen("data berhasil diubah"))
	}
}
