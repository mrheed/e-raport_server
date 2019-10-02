package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	db "github.com/syahidnurrohim/restapi/database"
	_ "github.com/syahidnurrohim/restapi/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MaterialStruct struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	KodeMateri  string             `json:"kode_materi,omitempty" bson:"kode_materi,omitempty"`
	NamaMapel   NamaMapel          `json:"nama_mapel,omitempty" bson:"nama_mapel,omitempty"`
	NamaMateri  string             `json:"nama_materi,omitempty" bson:"nama_materi,omitempty"`
	TahunAjaran int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

type NamaMapel struct {
	TahunAjaran int    `json:"tahun_ajaran" bson:"tahun_ajaran"`
	KodeMapel   string `json:"kode_mapel" bson:"kode_mapel"`
}

func NewMateri() *MaterialStruct {
	return &MaterialStruct{}
}

func (m *MaterialStruct) InsertMaterial(r *http.Request) error {
	var emptyStruct []interface{}
	err := json.NewDecoder(r.Body).Decode(&emptyStruct)
	if err != nil {
		return err
	}
	setting := NewSetting()
	appSetting, err := setting.GetAppSetting()
	if err != nil {
		return err
	}
	for _, k := range emptyStruct {
		dataMapel := k.(map[string]interface{})
		exist := findAndExist(db.Kompetensi, bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_materi": dataMapel["kode_materi"]})
		if exist {
			return errors.New("error: materi " + dataMapel["nama_materi"].(string) + " sudah tersedia")
		}
		dataMapel["tahun_ajaran"] = appSetting.TahunAjaran
		dataMapel["nama_mapel"] = map[string]interface{}{
			"kode_mapel":   dataMapel["nama_mapel"].(map[string]interface{})["value"],
			"tahun_ajaran": appSetting.TahunAjaran,
		}
	}
	_, err = db.Kompetensi.InsertMany(context.Background(), emptyStruct)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", setting)
	return nil
}

func (m *MaterialStruct) UpdateMaterial(filter map[string]interface{}, update map[string]interface{}) error {
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return err
	}
	delete(update, "_id")
	ID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
	if err != nil {
		return err
	}
	update["nama_mapel"] = bson.M{"kode_mapel": update["nama_mapel"].(map[string]interface{})["value"], "tahun_ajaran": appSetting.TahunAjaran}
	_, err = db.Kompetensi.UpdateOne(context.Background(), bson.M{"_id": ID, "tahun_ajaran": appSetting.TahunAjaran}, bson.M{"$set": update})
	if err != nil {
		return err
	}
	return nil
}

func (m *MaterialStruct) FindWithFilter(filter bson.M) ([]MaterialStruct, error) {
	var result []MaterialStruct
	cursor, err := db.Kompetensi.Find(context.Background(), filter)
	if err != nil {
		return []MaterialStruct{}, err
	}
	for cursor.Next(context.Background()) {
		var decoded MaterialStruct
		if err := cursor.Decode(&decoded); err != nil {
			continue
		}
		result = append(result, decoded)
	}
	return result, nil
}

func (m *MaterialStruct) GetAllMaterial() ([]bson.M, error) {
	var Material []bson.M
	setting := NewSetting()
	appSetting, err := setting.GetAppSetting()
	if err != nil {
		return []bson.M{}, err
	}
	cursor, err := db.Kompetensi.Find(context.Background(), bson.M{"tahun_ajaran": appSetting.TahunAjaran})
	if err != nil {
		return []bson.M{}, err
	}
	for cursor.Next(context.Background()) {
		var Kls MaterialStruct
		err := cursor.Decode(&Kls)
		if err != nil {
			return []bson.M{}, err
		}
		mapel := NewMapel()
		mapelData, err := mapel.GetSingleMapel(bson.M{
			"kode_mapel":   Kls.NamaMapel.KodeMapel,
			"tahun_ajaran": Kls.NamaMapel.TahunAjaran,
		})
		if err != nil {
			continue
		}
		mappedMaterial := bson.M{
			"_id":         Kls.ID,
			"kode_materi": Kls.KodeMateri,
			"nama_materi": Kls.NamaMateri,
			"nama_mapel": map[string]interface{}{
				"label": mapelData.NamaMapel,
				"value": Kls.NamaMapel.KodeMapel,
			},
			"tahun_ajaran": Kls.TahunAjaran,
		}
		Material = append(Material, mappedMaterial)
	}
	return Material, nil
}

func (m *MaterialStruct) DeleteMaterial(data []map[string]interface{}) error {
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return err
	}
	for _, f := range data {
		f["tahun_ajaran"] = appSetting.TahunAjaran
		fmt.Printf("%+v\n", f)
		if _, err := db.Kompetensi.DeleteOne(context.Background(), f); err != nil {
			return err
		}
	}
	return nil
}
