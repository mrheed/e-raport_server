package main

import (
	db "github.com/syahidnurrohim/restapi/database"
	r "github.com/syahidnurrohim/restapi/routehandler"
)

func main() {

	db.InitDBAndCollection()
	go r.CheckExpireTkn()
	r.InitRoute()
	defer db.Disconnect()
}
