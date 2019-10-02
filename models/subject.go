package models

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	db "github.com/syahidnurrohim/restapi/database"
	tool "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MapelStruct struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	KodeMapel     string             `json:"kode_mapel" bson:"kode_mapel"`
	NamaMapel     string             `json:"nama_mapel" bson:"nama_mapel"`
	KelompokMapel tool.CustomSelect  `json:"kelompok_mapel" bson:"kelompok_mapel"`
	MapelKelas    []MapelKelas       `json:"mapel_kelas" bson:"mapel_kelas"`
	TahunAjaran   int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

type MapelKelas struct {
	Jurusan     string `json:"jurusan" bson:"jurusan"`
	TahunAjaran int    `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

func NewMapel() *MapelStruct {
	return &MapelStruct{}
}

func (m *MapelStruct) GetRestructuredMapelWithJurusan(jurusan string) ([]tool.CustomSelect, error) {
	var result []tool.CustomSelect
	if tool.IsEmpty(jurusan) {
		return []tool.CustomSelect{}, errors.New("parameter jurusan tidak boleh kosong")
	}
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return []tool.CustomSelect{}, err
	}
	cursor, err := db.Mapel.Find(context.Background(), bson.M{"mapel_kelas": bson.M{"$elemMatch": jurusan}, "tahun_ajaran": appSetting.TahunAjaran})
	if err != nil {
		return []tool.CustomSelect{}, err
	}
	for cursor.Next(context.Background()) {
		var Mapel MapelStruct
		if err := cursor.Decode(&Mapel); err != nil {
			return []tool.CustomSelect{}, err
		}
		result = append(result, tool.CustomSelect{Value: Mapel.KodeMapel, Label: Mapel.NamaMapel})
	}
	return result, nil
}

func (m *MapelStruct) InsertMapel(data []bson.M) error {
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return err
	}
	for _, d1 := range data {
		d1["tahun_ajaran"] = appSetting.TahunAjaran
		mapelExist := findAndExist(db.Mapel, bson.M{"kode_mapel": d1["kode_mapel"], "tahun_ajaran": appSetting.TahunAjaran})
		if mapelExist {
			return errors.New("mapel " + d1["kode_mapel"].(string) + " sudah tersedia")
		}
		for i, mk1 := range d1["mapel_kelas"].([]interface{}) {
			mk1 := mk1.(map[string]interface{})
			mapelKelas, err := m.getMapelKelas(mk1["value"].(string))
			if err != nil {
				return err
			}
			d1["mapel_kelas"].([]interface{})[i] = mapelKelas
		}
		_, err = db.Mapel.InsertOne(context.Background(), d1)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MapelStruct) getMapelKelas(value string) (map[string]interface{}, error) {
	misc := NewMisc()
	classOnSchoolYear, err := misc.GetClassOnSchoolYear()
	if err != nil {
		return map[string]interface{}{}, err
	}
	spt := strings.Split(value, " ")
	tahunAjaran := classOnSchoolYear.SchoolYearOnGrade[spt[0]]
	return map[string]interface{}{
		"jurusan":      spt[1],
		"tahun_ajaran": tahunAjaran,
	}, nil
}

func (m *MapelStruct) reverseMapelKelas(mapelKelas MapelKelas) (bson.M, error) {
	classOnSchoolYear, err := NewMisc().GetClassOnSchoolYear()
	if err != nil {
		return bson.M{}, err
	}
	kelas := classOnSchoolYear.GradeOnSchoolYear[mapelKelas.TahunAjaran]
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return bson.M{}, err
	}
	jurusan, err := NewVocation().GetSingleVocation(bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_kelas": mapelKelas.Jurusan})
	if err != nil {
		return bson.M{}, err
	}
	result := bson.M{
		"value": kelas + " " + jurusan.KodeKelas,
		"label": kelas + " " + jurusan.NamaKelas,
	}
	return result, nil
}

func (m *MapelStruct) UpdateMapel(r *http.Request) error {
	var emptyStruct bson.M
	byteData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(byteData, &emptyStruct); err != nil {
		return err
	}
	updateState, assertTypeUpdate := emptyStruct["update"].(map[string]interface{})
	if !assertTypeUpdate {
		return errors.New("error type assertion")
	}
	delete(updateState, "_id")
	for _, d := range updateState["mapel_kelas"].([]interface{}) {
		d := d.(map[string]interface{})
		kodeKelas, isPassed := d["value"].(string)
		if !isPassed {
			continue
		}
		mapelKelas, err := m.getMapelKelas(kodeKelas)
		if err != nil {
			return err
		}
		d["jurusan"], d["tahun_ajaran"] = mapelKelas["jurusan"], mapelKelas["tahun_ajaran"]
	}
	filterState, assertTypeFilter := emptyStruct["filter"].(map[string]interface{})
	if !assertTypeFilter {
		return errors.New("error type assertion")
	}
	ID, err := primitive.ObjectIDFromHex(filterState["_id"].(string))
	_, err = db.Mapel.UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": updateState})
	if err != nil {
		return err
	}
	return nil
}

func (m *MapelStruct) GetAllMapel() ([]bson.M, error) {
	var result []bson.M
	setting := NewSetting()
	appSetting, err := setting.GetAppSetting()
	if err != nil {
		return []bson.M{}, err
	}
	cursor, err := db.Mapel.Find(context.Background(), bson.M{"tahun_ajaran": appSetting.TahunAjaran})
	if err != nil {
		return []bson.M{}, err
	}
	for cursor.Next(context.Background()) {
		var Mpl MapelStruct
		var tmpValue bson.M
		err := cursor.Decode(&Mpl)
		if err != nil {
			return []bson.M{}, err
		}
		tmpValue = bson.M{
			"kode_mapel":     Mpl.KodeMapel,
			"_id":            Mpl.ID,
			"nama_mapel":     Mpl.NamaMapel,
			"mapel_kelas":    []bson.M{},
			"tahun_ajaran":   Mpl.TahunAjaran,
			"kelompok_mapel": Mpl.KelompokMapel,
		}
		for _, mk := range Mpl.MapelKelas {
			reversed, err := m.reverseMapelKelas(mk)
			if err != nil {
				continue
			}
			tmpValue["mapel_kelas"] = append(tmpValue["mapel_kelas"].([]bson.M), reversed)
		}
		result = append(result, tmpValue)
	}
	return result, nil
}

func (m *MapelStruct) GetSingleMapel(filter bson.M) (MapelStruct, error) {
	var result MapelStruct
	err := db.Mapel.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		return MapelStruct{}, err
	}
	return result, nil
}

func (m *MapelStruct) DeleteMapel(r *http.Request) error {
	var filter []bson.M
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		return err
	}
	setting := NewSetting()
	appSetting, err := setting.GetAppSetting()
	if err != nil {
		return err
	}
	for _, f := range filter {
		f["tahun_ajaran"] = appSetting.TahunAjaran
		_, err = db.Mapel.DeleteOne(context.Background(), f)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MapelStruct) FindWithFilter(filter bson.M) ([]MapelStruct, error) {
	var result []MapelStruct
	cursor, err := db.Mapel.Find(context.Background(), filter)
	if err != nil {
		return []MapelStruct{}, err
	}
	for cursor.Next(context.Background()) {
		var decoded MapelStruct
		if err := cursor.Decode(&decoded); err != nil {
			continue
		}
		result = append(result, decoded)
	}
	return result, nil
}

func findAndExist(db *mongo.Collection, filter bson.M) bool {
	cursor, err := db.Find(context.Background(), filter)
	if err != nil {
		return false
	}
	return cursor.Next(context.Background())
}
