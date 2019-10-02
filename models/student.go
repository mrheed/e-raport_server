package models

import (
	"context"
	"encoding/json"
	"fmt"
	db "github.com/syahidnurrohim/restapi/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type Student struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NIS          int                `json:"nis,omitempty" bson:"nis,omitsempty"`
	Nama         string             `json:"nama,omitempty" bson:"nama,omitempty"`
	Jurusan      StudentJurusan     `json:"jurusan,omitempty" bson:"jurusan,omitempty"`
	JenisKelamin string             `json:"jeniskelamin,omitempty" bson:"jeniskelamin,omitempty"`
	TahunMasuk   int                `json:"tahun_masuk,omitempty" bson:"tahun_masuk,omitempty"`
}

type StudentJurusan struct {
	TahunAjaran int    `json:"tahun_ajaran" bson:"tahun_ajaran"`
	KodeKelas   string `json:"kode_kelas" bson:"kode_kelas"`
}

func NewStudent() *Student {
	return &Student{}
}

func (s *Student) GetAllStudents() ([]bson.M, error) {
	var result []bson.M
	var tahunMasuk []map[string]int
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return []bson.M{}, err
	}
	for _, d := range []int{appSetting.TahunAjaran, appSetting.TahunAjaran - 1, appSetting.TahunAjaran - 2} {
		tahunMasuk = append(tahunMasuk, map[string]int{"tahun_masuk": d})
	}
	cursor, err := db.Student.Find(context.Background(), bson.M{"$or": tahunMasuk})
	if err != nil {
		return []bson.M{}, err
	}
	for cursor.Next(context.Background()) {
		var tmpData Student
		if err := cursor.Decode(&tmpData); err != nil {
			return []bson.M{}, err
		}
		dataKelas, err := NewVocation().GetSingleVocation(bson.M{"tahun_ajaran": tmpData.Jurusan.TahunAjaran, "kode_kelas": tmpData.Jurusan.KodeKelas})
		if err != nil {
			continue
		}
		appendData := bson.M{
			"_id":          tmpData.ID,
			"nis":          tmpData.NIS,
			"nama":         tmpData.Nama,
			"jurusan":      bson.M{"label": dataKelas.NamaKelas, "value": dataKelas.KodeKelas},
			"jeniskelamin": tmpData.JenisKelamin,
			"tahun_masuk":  tmpData.TahunMasuk,
		}
		result = append(result, appendData)
	}
	return result, nil
}

func (s *Student) InsertStudents(r *http.Request) (int, error) {
	var emptyStruct []map[string]interface{}
	var insertSlice []interface{}
	if err := json.NewDecoder(r.Body).Decode(&emptyStruct); err != nil {
		return 0, err
	}
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return 0, err
	}
	for _, d := range emptyStruct {
		jurusan, ok := d["jurusan"].(map[string]interface{})
		if !ok {
			continue
		}
		d["jurusan"] = bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_kelas": jurusan["value"]}
		delete(d, "Jurusan")
		verified, err := s.VerifyStruct(d)
		if err != nil {
			fmt.Printf("%+v\n", err)
			continue
		}
		insertSlice = append(insertSlice, verified)
	}
	_, err = db.Student.InsertMany(context.Background(), insertSlice)
	if err != nil {
		return 0, err
	}
	return len(insertSlice), nil
}

func (s *Student) UpdateStudent(filter map[string]interface{}, update map[string]interface{}) error {
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return err
	}
	update["jurusan"] = bson.M{"kode_kelas": update["jurusan"].(map[string]interface{})["value"], "tahun_ajaran": appSetting.TahunAjaran}
	ID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
	if err != nil {
		return nil
	}
	filter["_id"] = ID
	verified, err := s.VerifyStruct(update)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", verified)
	_, err = db.Student.UpdateOne(context.Background(), filter, bson.M{"$set": verified})
	if err != nil {
		return err
	}
	return nil
}

func (s *Student) VerifyStruct(data map[string]interface{}) (Student, error) {
	var result Student
	byteData, err := json.Marshal(data)
	if err != nil {
		return Student{}, err
	}
	err = json.Unmarshal(byteData, &result)
	if err != nil {
		return Student{}, err
	}
	return result, nil
}

func (s *Student) FindWithFilter(filter bson.M) ([]Student, error) {
	var result []Student
	cursor, err := db.Student.Find(context.Background(), filter)
	if err != nil {
		return []Student{}, err
	}
	for cursor.Next(context.Background()) {
		var student Student
		err = cursor.Decode(&student)
		if err != nil {
			return []Student{}, err
		}
		result = append(result, student)
	}
	return result, nil
}

func (s *Student) GetSingleStudent(filter bson.M) (Student, error) {
	var result Student
	if err := db.Student.FindOne(context.Background(), filter).Decode(&result); err != nil {
		return Student{}, err
	}
	return result, nil
}
