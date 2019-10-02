package models

import (
	"context"
	"errors"
	"net/http"

	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Misc struct {
	Writer http.ResponseWriter
}

type DashboardRank struct {
	ID   int32         `json:"_id"`
	Data []primitive.M `json:"data"`
}

type ClassOnSchoolYear struct {
	GradeOnSchoolYear map[int]string `json:"grade_on_school_year"`
	SchoolYearOnGrade map[string]int `json:"school_year_on_grade"`
}

type DashboardInfo struct {
	TeacherCount int64           `json:"teacher_count"`
	StudentCount int64           `json:"student_count"`
	MapelCount   int64           `json:"mapel_count"`
	ClassCount   int64           `json:"class_count"`
	RankData     []DashboardRank `json:"rank_data"`
}

func NewMisc() *Misc {
	return &Misc{}
}

func (m *Misc) GetGradeList() []interface{} {
	return bson.A{"X", "XI", "XII"}
}

func (m *Misc) GetScoreInputTypes() []bson.M {
	return []bson.M{
		bson.M{"value": "UH", "label": "Ulangan Harian"},
		bson.M{"value": "PTS", "label": "Ulangan Tengah Semester"},
		bson.M{"value": "PAS", "label": "Ulangan Akhir Semester"},
	}
}

func (m *Misc) GetClassOnSchoolYear() (ClassOnSchoolYear, error) {
	var AppData Application
	err := db.AppSetting.FindOne(context.Background(), bson.D{}).Decode(&AppData)
	if err != nil {
		return ClassOnSchoolYear{}, errors.New(err.Error())
	}
	tahunAjaran := AppData.TahunAjaran
	gradeOnSchoolYear := map[int]string{
		tahunAjaran - 2: "XII",
		tahunAjaran - 1: "XI",
		tahunAjaran:     "X",
	}
	schoolYearOnGrade := make(map[string]int)
	for k, v := range gradeOnSchoolYear {
		schoolYearOnGrade[v] = k
	}
	return ClassOnSchoolYear{
		GradeOnSchoolYear: gradeOnSchoolYear,
		SchoolYearOnGrade: schoolYearOnGrade,
	}, nil
}

func (m *Misc) GetDashboardInfo(w http.ResponseWriter) (DashboardInfo, error) {
	var filteredResult []DashboardRank
	var result primitive.A
	var pipe []bson.M
	tool.ProcessPipeAggregate(tool.AgURI("DashboardRank"), &pipe, w)
	if !tool.ProcessDataAggregate(&result, w, db.Student, pipe) {
		return DashboardInfo{}, errors.New("tidak dapat mengambil data")
	}

	teacherCount, _ := db.Teacher.EstimatedDocumentCount(context.Background())
	studentCount, _ := db.Student.EstimatedDocumentCount(context.Background())
	mapelCount, _ := db.Mapel.EstimatedDocumentCount(context.Background())
	classCount, _ := db.Kelas.EstimatedDocumentCount(context.Background())

	for _, d := range result {
		var filteredData []primitive.M
		w, _ := d.(primitive.M)
		data, _ := w["data"].(primitive.A)
		for i, d := range data {
			dm := d.(primitive.M)
			for k, v := range dm {
				dm[k] = v
				dm["rank"] = i + 1
			}
			filteredData = append(filteredData, dm)
			if i == 9 {
				break
			}
		}

		filteredResult = append(filteredResult, DashboardRank{Data: filteredData, ID: w["_id"].(int32)})
	}

	return DashboardInfo{
		TeacherCount: teacherCount,
		StudentCount: studentCount,
		MapelCount:   mapelCount,
		ClassCount:   classCount,
		RankData:     filteredResult,
	}, nil
}
