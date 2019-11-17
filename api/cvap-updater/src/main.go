package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/lib/pq"
)

const gCvapUrl = "https://www2.census.gov/programs-surveys/decennial/rdo/datasets/2017/2017-cvap/CVAP_2013-2017_ACS_csv_files.zip"
const gSource = "Census"

func downloadDataZip() (*os.File, error) {
	var zipFile *os.File
	var err error

	resp, err := http.Get(gCvapUrl)
	if err != nil {
		goto done
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("Got status code %v", resp.StatusCode)
		goto done
	}
	zipFile, err = ioutil.TempFile("", "CVAP.zip")
	if err != nil {
		goto done
	}
	defer zipFile.Close()
	for {
		var buff [1024]byte
		n, err := resp.Body.Read(buff[:])
		if n > 0 {
			_, err2 := zipFile.Write(buff[:n])
			if err2 != nil {
				err = err2
				goto done
			}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
				break
			}
			goto done
		}
	}
	_, err = zipFile.Seek(0, 0)
	if err != nil {
		goto done
	}

done:
	if err != nil && zipFile != nil {
		zipFile.Close()
		os.Remove(zipFile.Name())
		zipFile = nil
	}
	return zipFile, err
}

func getDistrictDataFromZip(zipFile *os.File) (*os.File, error) {
	var dataFile *os.File
	var err error
	foundCdCsv := false
	var fileStat os.FileInfo
	var zipReader *zip.Reader

	// look for CD.csv
	fileStat, err = zipFile.Stat()
	if err != nil {
		goto done
	}
	zipReader, err = zip.NewReader(zipFile, fileStat.Size())
	if err != nil {
		goto done
	}
	dataFile, err = ioutil.TempFile("", "CD.csv")
	if err != nil {
		goto done
	}
	for _, entry := range zipReader.File {
		if entry.Name != "CD.csv" {
			continue
		}

		foundCdCsv = true
		f, err := entry.Open()
		if err != nil {
			goto done
		}
		defer f.Close()
		if _, err = io.Copy(dataFile, f); err != nil {
			goto done
		}
		break
	}
	if !foundCdCsv {
		err = fmt.Errorf("Didn't find CD.csv")
		goto done
	}

	_, err = dataFile.Seek(0, 0)
	if err != nil {
		goto done
	}

done:
	if err != nil && dataFile != nil {
		dataFile.Close()
		os.Remove(dataFile.Name())
		dataFile = nil
	}
	return dataFile, err
}

func parseCongressNbr(geoname string) *int {
	/*
		Format: "Congressional District 1 (115th Congress), Alabama"
	*/

	// find number string
	parts := strings.Fields(geoname)
	congressPos := -1
	for i, part := range parts {
		if strings.Contains(part, "Congress)") {
			congressPos = i
			break
		}
	}
	if congressPos == -1 || congressPos == 0 {
		log.Printf("Here 1: %v %v %v", congressPos, len(parts), parts[congressPos])
		return nil
	}

	// get number string
	nbrPos := congressPos - 1
	var nbrStrBuilder strings.Builder
	toParse := parts[nbrPos]
	if toParse[0] == '(' {
		toParse = toParse[1:]
	}
	for _, c := range toParse {
		if unicode.IsDigit(c) {
			nbrStrBuilder.WriteRune(c)
		} else {
			break
		}
	}
	nbrStr := nbrStrBuilder.String()
	if len(nbrStr) == 0 {
		log.Printf("Here 2")
		return nil
	}

	// parse number string
	n, err := strconv.Atoi(nbrStr)
	if err != nil {
		log.Printf("Here 3")
		return nil
	}
	return &n
}

type geoid struct {
	state    int
	district string
}

func parseGeoid(s string) *geoid {
	/*
		Format: "50000USssdd", where ss = two-digit state FIPS code and dd = the two-digit district number
	*/

	const prefix = "50000US"
	if !strings.HasPrefix(s, prefix) {
		return nil
	}
	if len(s) != len(prefix)+4 {
		return nil
	}
	var g geoid
	var err error
	g.state, err = strconv.Atoi(s[len(prefix) : len(prefix)+2])
	if err != nil {
		return nil
	}
	g.district = s[len(prefix)+2:]
	return &g
}

func parseNumber(s string) *int {
	s = strings.TrimSpace(s)
	var n int
	if len(s) == 0 {
		n = 0
	} else {
		var err error
		n, err = strconv.Atoi(s)
		if err != nil {
			log.Printf("Failed to parse number: %v", err)
			return nil
		}
	}
	return &n
}

func updateDatabase(ctx context.Context, dataFile *os.File, stateData map[int]string) error {
	// connect to DB
	connStr := "host=db user=postgres password=pw dbname=house sslmode=disable connect_timeout=10"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// delete old data
	log.Printf("Deleting old data")
	_, err = db.ExecContext(ctx, "DELETE FROM source WHERE name = $1", gSource)
	if err != nil {
		return err
	}

	// make new source row
	result, err := db.QueryContext(ctx, "INSERT INTO source(name) VALUES ($1) RETURNING id", gSource)
	if err != nil {
		return err
	}
	defer result.Close()
	result.Next()
	var sourceId string
	if err := result.Scan(&sourceId); err != nil {
		return err
	}

	// read data
	sql := "INSERT INTO house_district(state, district, congress, pop," +
		" pop_moe, cvap, cvap_moe, pop_source, cvap_source) VALUES" +
		" ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	reader := csv.NewReader(dataFile)
	reader.ReuseRecord = true
	nbrInserted, nbrError := 0, 0
	for {
		rec, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Printf("Parsing error: %v", err)
			nbrError++
			continue
		}
		if len(rec) == 0 || rec[0] == "GEONAME" {
			continue
		}

		// check LNNUMBER
		lnnumber := strings.TrimSpace(rec[3])
		if lnnumber != "1" {
			continue
		}

		// get data
		congressNbr := parseCongressNbr(rec[0])
		if congressNbr == nil {
			log.Printf("Failed to parse Congress number")
			nbrError++
			continue
		}
		g := parseGeoid(strings.TrimSpace(rec[2]))
		if g == nil {
			log.Printf("Failed to parse GEOID: %v", rec[2])
			nbrError++
			continue
		}
		pop := parseNumber(rec[4])
		if pop == nil {
			nbrError++
			continue
		}
		popMoe := parseNumber(rec[5])
		if pop == nil {
			nbrError++
			continue
		}
		cvap := parseNumber(rec[10])
		if pop == nil {
			nbrError++
			continue
		}
		cvapMoe := parseNumber(rec[11])
		if pop == nil {
			nbrError++
			continue
		}
		state, ok := stateData[g.state]
		if !ok {
			log.Printf("Invalid state code: %v", g.state)
			nbrError++
			continue
		}

		// insert into DB
		_, err = db.ExecContext(ctx, sql, state, g.district,
			congressNbr, *pop, *popMoe, *cvap, *cvapMoe, sourceId, sourceId)
		if err != nil {
			return err
		}

		nbrInserted++
	}

	log.Printf("Inserted %v records (had %v errors)", nbrInserted, nbrError)
	return nil
}

func handleSignals(f func()) {
	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)
	go func() {
		select {
		case <-appSignal:
			log.Printf("Got signal")
			f()
		}
	}()
}

type stateDataEntry struct {
	Name string
	FIPS int
	USPS string
}

// Returns a map from state FIPS code to state abbrev
func loadStateData(statesFilePath string) (map[int]string, error) {
	f, err := os.Open(statesFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	states := make(map[int]string)
	if !decoder.More() {
		return nil, errors.New("States file seems to be empty")
	}
	var entries []stateDataEntry
	err = decoder.Decode(&entries)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		states[entry.FIPS] = entry.USPS
	}

	return states, nil
}

func main() {
	// parse args
	var dataZipPath *string = flag.String("datazip", "", "zip file containing data")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	statesFilePath := flag.Arg(0)

	// load state data
	stateData, err := loadStateData(statesFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// download/open zipfile
	var dataZip *os.File
	if len(*dataZipPath) == 0 {
		// download CVAP data
		log.Print("Downloading data")
		dataZip, err = downloadDataZip()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Print("Found already-downloaded data")
		dataZip, err = os.Open(*dataZipPath)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer dataZip.Close()

	// get district data from zipfile
	districtData, err := getDistrictDataFromZip(dataZip)
	if err != nil {
		log.Fatal(err)
	}
	defer districtData.Close()
	log.Print("Got district data")

	// update DB
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	handleSignals(stop)
	err = updateDatabase(ctx, districtData, stateData)
	if err != nil {
		log.Fatal(err)
	}
}
