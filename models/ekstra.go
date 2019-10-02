package models

import (
	"context"
	db "github.com/syahidnurrohim/restapi/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EkstraStruct struct {
	ID          primitive.ObjectID `json:"_id, omitempty" bson:"_id, omitempty"`
	KodeEkstra  string             `json:"kode_ekstra, omitempty" bson:"kode_ekstra, omitempty"`
	NamaEkstra  string             `json:"nama_ekstra" bson:"nama_ekstra"`
	Pembimbing  string             `json:"pembimbing" bson:"pembimbing"`
	TahunAjaran int                `json:"tahun_ajaran" bson:"tahun_ajaran"`
}

func NewEkstra() *EkstraStruct {
	return &EkstraStruct{}
}

func (e *EkstraStruct) GetEkstraData() ([]EkstraStruct, error) {
	var result []EkstraStruct
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return []EkstraStruct{}, err
	}
	cursor, err := db.Ekstra.Find(context.Background(), bson.M{"tahun_ajaran": appSetting.TahunAjaran})
	if err != nil {
		return []EkstraStruct{}, err
	}
	for cursor.Next(context.Background()) {
		var tmp EkstraStruct
		if err := cursor.Decode(&tmp); err != nil {
			continue
		}
		result = append(result, tmp)
	}
	return result, nil
}

func (e *EkstraStruct) findAndExist(filter bson.M) bool {
	var result EkstraStruct
	if err := db.Ekstra.FindOne(context.Background(), filter).Decode(&result); err != nil {
		return true
	}
	if result == (EkstraStruct{}) {
		return true
	}
	return false
}

func (e *EkstraStruct) marshal(bson.M) (EkstraStruct, error) {
	return EkstraStruct{}, nil
}

func (e *EkstraStruct) InsertEkstraData(data []EkstraStruct) error {
	appSetting, err := NewSetting().GetAppSetting()
	if err != nil {
		return err
	}
	for _, d := range data {
		if e.findAndExist(bson.M{"tahun_ajaran": appSetting.TahunAjaran, "kode_ekstra": d.KodeEkstra}) {
			continue
		}
		if _, err := db.Ekstra.InsertOne(context.Background(), d); err != nil {
			return err
		}
	}
	return nil
}

func (e *EkstraStruct) UpdateEkstraData(filter bson.M, update bson.M) error {
	ID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
	if err != nil {
		return err
	}
	updateVal, err := e.marshal(update)
	if err != nil {
		return err
	}
	if _, err := db.Ekstra.UpdateOne(context.Background(), bson.M{"_id": ID}, bson.M{"$set": updateVal}); err != nil {
		return err
	}
	return nil
}

func (e *EkstraStruct) DeleteEkstraData(filter []bson.M) error {
	if _, err := db.Ekstra.DeleteMany(context.Background(), bson.M{"$or": filter}); err != nil {
		return err
	}
	return nil
}
