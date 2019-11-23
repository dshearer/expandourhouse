package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var gDb *sql.DB

type entry struct {
	Congress int
	Pop      int
	PopMoe   int
	Cvap     int
	CvapMoe  int
}

func handleGetDistricts(resp http.ResponseWriter, req *http.Request) {
	encoder := json.NewEncoder(resp)

	// load data from DB
	data := make(map[string]entry)
	rows, err := gDb.QueryContext(req.Context(), 
		"SELECT state, district, congress, pop, pop_moe, cvap, cvap_moe FROM house_district")
	if err != nil {
		goto done
	}
	defer rows.Close()
	for rows.Next() {
		var e entry
		var state string
		var district string
		if err = rows.Scan(&state, &district, &e.Congress, &e.Pop,
			&e.PopMoe, &e.Cvap, &e.CvapMoe); err != nil {
			goto done
		}
		data[fmt.Sprintf("%v%v", state, district)] = e
	}

	// make response
	encoder.Encode(data)

done:
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
}

func main() {
	// connect to DB
	var err error
	connStr := "host=db user=postgres password=pw dbname=house sslmode=disable connect_timeout=10"
	gDb, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer gDb.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/districts", handleGetDistricts).Methods("GET")
	srv := &http.Server{
		Handler: r,
		Addr:    ":80",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
