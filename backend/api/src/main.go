package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var gDb *sql.DB

const gContentTypeHeader = "Content-Type"
const gJSONContentType = "application/json"

type congressInfo struct {
	StartYear int `json:"startYear"`
}

func handleGetCongresses(resp http.ResponseWriter, req *http.Request) {
	congresses := make(map[string]*congressInfo)

	// get congresses from DB
	rows, err := gDb.QueryContext(req.Context(),
		"SELECT nbr, start_year FROM congress")
	if err != nil {
		goto done
	}
	defer rows.Close()
	for rows.Next() {
		var nbr int
		var con congressInfo
		if err = rows.Scan(&nbr, &con.StartYear); err != nil {
			goto done
		}
		congresses[fmt.Sprintf("%v", nbr)] = &con
	}

done:
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}

	// make response
	resp.Header().Add(gContentTypeHeader, gJSONContentType)
	json.NewEncoder(resp).Encode(congresses)
}

type fact struct {
	Value  int    `json:"value"`
	Source string `json:"source"`
}

type factWithMoe struct {
	fact
	MarginOfError int `json:"marginOfError"`
}

type districtFacts map[string]interface{}

type stateInfo struct {
	IrregularHow []string                 `json:"irregularHow"`
	Districts    map[string]districtFacts `json:"districts"`
}

func handleGetStates(resp http.ResponseWriter, req *http.Request) {
	var err error
	statusCode := http.StatusInternalServerError
	var rows *sql.Rows
	var states []string
	var allDistricts map[string][]*districtInfo
	result := make(map[string]*stateInfo)

	// get vars
	vars := mux.Vars(req)
	congress, err := strconv.Atoi(vars["congress"])
	if err != nil {
		statusCode = http.StatusBadRequest
		goto done
	}

	// get all states for this congress
	states, err = getStates(req.Context(), congress)
	if err != nil {
		goto done
	}

	// get all districts for this congress
	allDistricts, err = getDistricts(req.Context(), congress)
	if err != nil {
		goto done
	}

	// get info for each state
	for _, stateAbbr := range states {
		state := stateInfo{IrregularHow: nil, Districts: nil}
		result[stateAbbr] = &state

		// look for irregularities
		state.IrregularHow, err = getStateIrregularities(req.Context(), stateAbbr, congress)
		if err != nil {
			goto done
		}
		if len(state.IrregularHow) > 0 {
			continue
		}

		// get district info
		state.Districts = make(map[string]districtFacts)
		for _, di := range allDistricts[stateAbbr] {
			distFacts, err := getDistrictFacts(req.Context(), di.rowID)
			if err != nil {
				goto done
			}
			state.Districts[fmt.Sprintf("%v", di.nbr)] = distFacts
		}
	}

done:
	if rows != nil {
		rows.Close()
	}
	if err != nil {
		resp.WriteHeader(statusCode)
		log.Print(err)
		return
	}

	// make response
	resp.Header().Add(gContentTypeHeader, gJSONContentType)
	json.NewEncoder(resp).Encode(result)
}

func handleGetDistrict(resp http.ResponseWriter, req *http.Request) {
	var err error
	statusCode := http.StatusInternalServerError
	var rows *sql.Rows
	var congress int
	var state string
	var district int
	var districtID *int
	var result districtFacts

	// get vars
	vars := mux.Vars(req)
	congress, err = strconv.Atoi(vars["congress"])
	if err != nil {
		statusCode = http.StatusBadRequest
		goto done
	}
	state = vars["state"]
	district, err = strconv.Atoi(vars["district"])
	if err != nil {
		statusCode = http.StatusBadRequest
		goto done
	}

	// get facts
	districtID, err = getDistrictID(req.Context(), congress, state, district)
	if err != nil {
		goto done
	}
	if districtID == nil {
		statusCode = http.StatusNotFound
		goto done
	}
	result, err = getDistrictFacts(req.Context(), *districtID)
	if err != nil {
		goto done
	}

done:
	if rows != nil {
		rows.Close()
	}
	if err != nil {
		resp.WriteHeader(statusCode)
		if err != nil {
			log.Print(err)
		}
		return
	}

	// make response
	resp.Header().Add(gContentTypeHeader, gJSONContentType)
	json.NewEncoder(resp).Encode(result)
}

func main() {
	// connect to DB
	var err error
	connStr := "host=db user=postgres password=pw dbname=house" +
		" sslmode=disable connect_timeout=10"
	gDb, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer gDb.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/congresses", handleGetCongresses).Methods("GET")
	r.HandleFunc("/api/congresses/{congress}/states", handleGetStates).Methods("GET")
	r.HandleFunc("/api/congresses/{congress}/states/{state}/districts/{district}",
		handleGetDistrict).Methods("GET")
	srv := &http.Server{
		Handler:      r,
		Addr:         ":80",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
