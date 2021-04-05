package routehandler

import (
  "net/http"
  "strconv"

  "github.com/gorilla/mux"
  mod "github.com/syahidnurrohim/restapi/models"
  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
)

func GetPrintController(w http.ResponseWriter, r *http.Request) {
  var result primitive.A
  var pipe []bson.M
  tipe := mux.Vars(r)["type"]
  section := mux.Vars(r)["section"]
  throw := mod.NewThrower(w)
  reporter := mod.NewReporter(&pipe, &result, w)
  throw.StatusCode = http.StatusMovedPermanently
  nis := tool.GQry("nis", r)
  tayp := tool.GQry("tipe", r)
  mapel := tool.GQry("mapel", r)
  materi := tool.GQry("materi", r)
  jurusan := tool.GQry("jurusan", r)
  namaTugas := tool.GQry("nama_tugas", r)
  tahunMasuk := tool.GQry("tahun_masuk", r)
  tahunAjaran := tool.GQry("tahun_ajaran", r)
  switch tipe {
  case "assignment", "result_rapor", "UH", "PTS", "PAS":
    switch section {
    case "student":
      tahunMasuk, _ := strconv.Atoi(tahunMasuk)
      if tool.IsEmpty(tahunMasuk) {
        throw.Error("invalid tahun_masuk value")
        return
      }
      if err := reporter.GetReportStudent(jurusan, tahunMasuk); err != nil {
        throw.Error(err.Error())
        return
      }
      throw.Response(&result)
    case "subject":
      if tool.IsEmpty(jurusan) {
        throw.Error("parameter jurusan cannot be empty")
        return
      }
      if tool.IsEmpty(tahunAjaran) {
        throw.Error("parameter tahun_ajaran cannot be empty")
        return
      }
      tahunAjaran, err := strconv.Atoi(tahunAjaran)
      if err != nil {
        throw.Error(err.Error())
        return
      }
      if err := reporter.GetReportSubject(jurusan, tahunAjaran); err != nil {
        throw.Error(err.Error())
        return
      }
      throw.Response(&result)
    case "material":
      if tool.IsEmpty(mapel) {
        throw.Error("parameter mapel cannot be empty")
        return
      }
      if err := reporter.GetReportMaterial(mapel); err != nil {
        throw.Error(err.Error())
        return
      }
      throw.Response(&result)
    case "task_name":
      if tool.IsEmpty(mapel) {
        throw.Error("parameter mapel cannot be empty")
        return
      }
      if err := reporter.GetReportTaskName(mapel); err != nil {
        throw.Error(err.Error())
        return
      }
      throw.Response(&result)
    case "exam_result":
      if tool.IsEmpty(nis) {
        throw.Error("parameter nis cannot be empty")
        return
      } else if tayp == "UH" && tool.IsEmpty(materi) {
        throw.Error("parameter materi cannot be empty")
        return
      } else if tool.IsEmpty(mapel) {
        throw.Error("parameter mapel cannot be empty")
        return
      } else if tool.IsEmpty(jurusan) {
        throw.Error("parameter jurusan cannot be empty")
        return
      } else if tool.IsEmpty(tayp) {
        throw.Error("parameter tipe cannot be empty")
        return
      } else if err := reporter.GetReportExamResult(nis, materi, mapel, jurusan, tayp); err != nil {
        throw.Error(err.Error())
        return
      } else {
        throw.Response(&result)
      }
    case "task_result":
      if tool.IsEmpty(nis) {
        throw.Error("parameter nis cannot be empty")
        return
      } else if tool.IsEmpty(namaTugas) {
        throw.Error("parameter nama_tugas cannot be empty")
        return
      } else if tool.IsEmpty(mapel) {
        throw.Error("parameter mapel cannot be empty")
        return
      } else if err := reporter.GetReportTaskResult(nis, namaTugas, mapel, jurusan); err != nil {
        throw.Error(err.Error())
        return
      } else {
        throw.Response(&result)
      }
    case "final_result":

    default:
      throw.Error("invalid section parameter")
    }
  default:
    throw.Error("invalid type parameter")
  }
}
