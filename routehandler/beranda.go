package routehandler

import (
  "net/http"

  tool "github.com/syahidnurrohim/restapi/utils"
)

func GetDashboardCount(w http.ResponseWriter, r *http.Request) {
  tool.HeadersHandler(w)
  // if _, err := tool.VerifyHeader("Authorization", r, w); err == nil {

  // }
}
