package routehandler

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type InitServer struct {
	r *mux.Router
}

// InitRoute , all route will be placed here
func InitRoute() {

	Router := mux.NewRouter()
	// Nilai UH route
	Router.HandleFunc("/api/misc/{types}", GetMiscTypes).Methods("GET")
	// Cetak route
	Router.HandleFunc("/api/cetak/{type}/{section}", GetPrintController).Methods("GET")
	// Nilai PTPAS route
	Router.HandleFunc("/api/nilai/{type}", InsertScoreController).Methods("POST")
	Router.HandleFunc("/api/nilai/{type}", GetScoreController).Methods("GET")
	Router.HandleFunc("/api/nilai/{type}", UpdateScoreController).Methods("PUT")
	// Teacher Route
	Router.HandleFunc("/api/teacher/{id}", GetTeacherController).Methods("GET")
	Router.HandleFunc("/api/teacher", GetTeachersController).Methods("GET")
	Router.HandleFunc("/api/teacher", InsertTeacherController).Methods("POST")
	Router.HandleFunc("/api/teacher", UpdateDeleteTeacherController).Methods("PUT", "DELETE", "OPTIONS")
	// User Route
	Router.HandleFunc("/api/user/{id}", GetUserController).Methods("GET", "OPTIONS")
	Router.HandleFunc("/api/user/{id}", DeleteUserController).Methods("DELETE", "OPTIONS")
	Router.HandleFunc("/api/user/{id}", EditUserController).Methods("PUT", "OPTIONS")
	Router.HandleFunc("/api/user", GetUsersController).Methods("GET", "OPTIONS")
	Router.HandleFunc("/api/user", CreateUserController).Methods("POST", "OPTIONS")
	// Student Route
	Router.HandleFunc("/api/student", GetStudentsController).Methods("GET")
	Router.HandleFunc("/api/student", InsertStudentController).Methods("POST")
	Router.HandleFunc("/api/student", UpdateDeleteStudentController).Methods("PUT", "DELETE", "OPTIONS")
	// Ekstra Route
	Router.HandleFunc("/api/ekstra", GetEkstraController).Methods("GET")
	Router.HandleFunc("/api/ekstra", InsertEkstraController).Methods("POST")
	Router.HandleFunc("/api/ekstra", UpdateDeleteEkstraController).Methods("PUT", "DELETE", "OPTIONS")
	// Mapel Route
	Router.HandleFunc("/api/mapel", GetSubjectsController).Methods("GET")
	Router.HandleFunc("/api/mapel/{id}", GetSubjectController).Methods("GET")
	Router.HandleFunc("/api/mapel", InsertSubjectController).Methods("POST")
	Router.HandleFunc("/api/mapel", UpdateDeleteSubjectController).Methods("PUT", "DELETE", "OPTIONS")
	// Materi Route
	Router.HandleFunc("/api/kd", GetCompetenciesController).Methods("GET")
	Router.HandleFunc("/api/kd/{id}", GetCompetenceController).Methods("GET")
	Router.HandleFunc("/api/kd", InsertCompetenceController).Methods("POST")
	Router.HandleFunc("/api/kd", UpdateDeleteCompetenceController).Methods("PUT", "DELETE", "OPTIONS")
	// Kelas Route
	Router.HandleFunc("/api/kelas", GetClassesController).Methods("GET")
	Router.HandleFunc("/api/kelas/{misc}", GetClassController).Methods("GET")
	Router.HandleFunc("/api/kelas", InsertClassController).Methods("POST")
	Router.HandleFunc("/api/kelas", UpdateDeleteClassController).Methods("PUT", "DELETE", "OPTIONS")
	// Authentication Route
	Router.HandleFunc("/api/refresh", Refresh).Methods("POST", "OPTIONS")
	Router.HandleFunc("/api/purge", PurgeSingleToken).Methods("POST", "OPTIONS")
	Router.HandleFunc("/api/auth", SignInRequestHandler).Methods("POST", "OPTIONS")
	// App Setting Route
	Router.HandleFunc("/api/setting/{tab}/{type}", UpdateSettingController).Methods("PUT", "OPTIONS")
	Router.HandleFunc("/api/setting/{tab}", InsertSettingController).Methods("POST", "OPTIONS")
	Router.HandleFunc("/api/setting", GetSettingController).Methods("GET")

	http.Handle("/", &InitServer{Router})
	log.Println("Starting...")

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func (s *InitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		log.Println(origin)
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}

	if r.Method == "OPTIONS" {
		return
	}
	s.r.ServeHTTP(w, r)
}
