package routehandler

import (
  "context"
  "encoding/json"
  "io/ioutil"
  "log"
  "net/http"
  "path/filepath"
  "strconv"
  "strings"

  "github.com/gorilla/mux"
  mod "github.com/syahidnurrohim/restapi/models"
  tool "github.com/syahidnurrohim/restapi/utils"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
)

func GetMiscTypes(w http.ResponseWriter, r *http.Request) {
  throw := mod.NewThrower(w)
  throw.StatusCode = http.StatusMovedPermanently
  if _, err := tool.VerifyHeader(tool.GetRole().AllUser, r, w); err != nil {
  }
  misc := mod.NewMisc()
  switch mux.Vars(r)["types"] {
  case "extype":
    throw.Response(misc.GetScoreInputTypes())
  case "grades":
    throw.Response(misc.GetGradeList())
  case "class_on_school_year":
    data, err := misc.GetClassOnSchoolYear()
    if err != nil {
      throw.Error(err.Error())
      return
    }
    log.Println(data)
    throw.Response(&data)
  case "dashboard_info":
    data, err := misc.GetDashboardInfo(w)
    if err != nil {
      throw.Error(err.Error())
      return
    }
    throw.Response(&data)
  default:
    throw.Error("invalid type parameters")
  }
}

func appendCommaToMap(keyName string, r *http.Request) []byte {
  var nisSlice []map[string]int
  for _, s := range strings.Split(tool.GQry("siswa", r), ",") {
    nis, _ := strconv.Atoi(s)
    dt := map[string]int{
      keyName: nis,
    }
    nisSlice = append(nisSlice, dt)
  }
  mb, _ := json.Marshal(nisSlice)
  return mb
}

func gQry(key string, r *http.Request) string {
  return r.URL.Query().Get(key)
}

func processDataAggregate(result *primitive.A, w http.ResponseWriter, collection *mongo.Collection, pipe []bson.M) bool {
  cursor, err := collection.Aggregate(context.Background(), pipe)
  if err != nil {
    http.Error(w, tool.JSONErr(err.Error()), http.StatusInternalServerError)
    return false
  }
  for cursor.Next(context.Background()) {
    var elem bson.D
    cursor.Decode(&elem)
    *result = append(*result, elem.Map())
  }
  return true
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
