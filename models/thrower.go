package models

import (
  "encoding/json"
  "net/http"

  tool "github.com/syahidnurrohim/restapi/utils"
)

type Thrower struct {
  StatusCode int
  Writer     http.ResponseWriter
}

func NewThrower(w http.ResponseWriter) *Thrower {
  return &Thrower{
    Writer: w,
  }
}

func (e *Thrower) Error(errMsg string) {
  http.Error(e.Writer, tool.JSONErr(errMsg), e.StatusCode)
}

func (e *Thrower) Response(response interface{}) {
  json.NewEncoder(e.Writer).Encode(&response)
}
