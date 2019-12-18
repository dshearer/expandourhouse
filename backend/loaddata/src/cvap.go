package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"unicode"
)

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
	district int
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
	g.district, err = strconv.Atoi(s[len(prefix)+2:])
	if err != nil {
		return nil
	}
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

type surveyData struct {
	congressNbr *int
	g           *geoid
	pop         *int
	popMoe      *int
	adults      *int
	adultsMoe   *int
	citizens    *int
	citizensMoe *int
	cvap        *int
	cvapMoe     *int
}

type surveyReader struct {
	csvReader    *csv.Reader
	colNameToIdx map[string]int
}

func newSurveyReader(f *os.File) *surveyReader {
	r := &surveyReader{csv.NewReader(f), nil}
	r.csvReader.ReuseRecord = true
	return r
}

func (self *surveyReader) read() (*surveyData, error) {
	var rec []string
	var err error
	var data surveyData

do:
	rec, err = self.csvReader.Read()
	if err != nil {
		return nil, err
	}
	if len(rec) == 0 {
		goto do
	}

	if self.colNameToIdx == nil {
		// keep column names
		self.colNameToIdx = make(map[string]int)
		for idx, colName := range rec {
			self.colNameToIdx[colName] = idx
		}
		goto do
	}

	lnNbrColIdx, hasLnNbr := self.colNameToIdx["LNNUMBER"]
	geoNameColIdx, hasGeoName := self.colNameToIdx["GEONAME"]
	geoIDColIdx, hasGeoID := self.colNameToIdx["GEOID"]
	totEstColIdx, hasTotEst := self.colNameToIdx["TOT_EST"]
	totMoeColIdx, hasTotMoe := self.colNameToIdx["TOT_MOE"]
	aduEstColIdx, hasAduEst := self.colNameToIdx["ADU_EST"]
	aduMoeColIdx, hasAduMoe := self.colNameToIdx["ADU_MOE"]
	citEstColIdx, hasCitEst := self.colNameToIdx["CIT_EST"]
	citMoeColIdx, hasCitMoe := self.colNameToIdx["CIT_MOE"]
	cvapEstColIdx, hasCvapEst := self.colNameToIdx["CVAP_EST"]
	cvapMoeColIdx, hasCvapMoe := self.colNameToIdx["CVAP_MOE"]

	// check LNNUMBER
	if hasLnNbr {
		lnnumber := strings.TrimSpace(rec[lnNbrColIdx])
		if lnnumber != "1" {
			goto do
		}
	}

	if hasGeoName {
		data.congressNbr = parseCongressNbr(rec[geoNameColIdx])
		if data.congressNbr == nil {
			log.Printf("Failed to parse Congress number")
			return nil, err
		}
	}

	if hasGeoID {
		data.g = parseGeoid(strings.TrimSpace(rec[geoIDColIdx]))
		if data.g == nil {
			log.Printf("Failed to parse GEOID: %v", rec[geoIDColIdx])
			return nil, err
		}
	}

	if hasTotEst {
		data.pop = parseNumber(rec[totEstColIdx])
	}

	if hasTotMoe {
		data.popMoe = parseNumber(rec[totMoeColIdx])
	}

	if hasAduEst {
		data.adults = parseNumber(rec[aduEstColIdx])
	}

	if hasAduMoe {
		data.adultsMoe = parseNumber(rec[aduMoeColIdx])
	}

	if hasCitEst {
		data.citizens = parseNumber(rec[citEstColIdx])
	}

	if hasCitMoe {
		data.citizensMoe = parseNumber(rec[citMoeColIdx])
	}

	if hasCvapEst {
		data.cvap = parseNumber(rec[cvapEstColIdx])
	}

	if hasCvapMoe {
		data.cvapMoe = parseNumber(rec[cvapMoeColIdx])
	}

	return &data, nil
}

func addDistrictPop(ctx context.Context, db *sql.DB, districtRowId int,
	typ string, value, moe, sourceId int) error {

	var rows *sql.Rows
	var err error
	var exists bool
	var sql string

	// check if already exits
	sql = "SELECT COUNT(*) FROM house_district_pop WHERE" +
		" house_district_id = $1 AND type = $2"
	rows, err = db.QueryContext(ctx, sql, districtRowId, typ)
	if err != nil {
		goto done
	}
	rows.Next()
	if err = rows.Scan(&exists); err != nil {
		goto done
	}
	if exists {
		goto done
	}

	// add row
	sql = "INSERT INTO house_district_pop(house_district_id, type, value, " +
		"margin_of_error, source_id) VALUES ($1, $2, $3, $4, $5)"
	_, err = db.ExecContext(ctx, sql, districtRowId, typ, value, moe, sourceId)

done:
	if rows != nil {
		rows.Close()
	}
	return err
}

func updateDatabase(ctx context.Context, db *sql.DB,
	dataFile *os.File, stateData map[int]string) error {

	sourceText := "US Census Citizen Voting Age Population by Race and Ethnicity 2013-2017"
	sourceId, err := getSource(ctx, db, sourceText)
	if err != nil {
		return err
	}

	// read data
	reader := newSurveyReader(dataFile)
	nbrInserted, nbrError := 0, 0
	for {
		rec, err := reader.read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if rec.congressNbr == nil {
			continue
		}

		// look up state
		state, ok := stateData[rec.g.state]
		if !ok {
			log.Printf("Invalid state code: %v", rec.g.state)
			nbrError++
			continue
		}

		// get district row ID
		districtRowId, err := getDistrict(ctx, db, state, rec.g.district, *rec.congressNbr)
		if err != nil {
			return err
		}

		if rec.pop != nil && rec.popMoe != nil {
			err = addDistrictPop(ctx, db, districtRowId, "all", *rec.pop,
				*rec.popMoe, sourceId)
			if err != nil {
				return err
			}
		}
		if rec.adults != nil && rec.adultsMoe != nil {
			err = addDistrictPop(ctx, db, districtRowId, "adults",
				*rec.adults, *rec.adultsMoe, sourceId)
			if err != nil {
				return err
			}
		}
		if rec.citizens != nil && rec.citizensMoe != nil {
			err = addDistrictPop(ctx, db, districtRowId, "citizens", *rec.citizens,
				*rec.citizensMoe, sourceId)
			if err != nil {
				return err
			}
		}
		if rec.cvap != nil && rec.cvapMoe != nil {
			err = addDistrictPop(ctx, db, districtRowId, "cvap", *rec.cvap,
				*rec.cvapMoe, sourceId)
			if err != nil {
				return err
			}
		}

		nbrInserted++
	}

	log.Printf("Inserted %v records (had %v errors)", nbrInserted, nbrError)
	return nil
}

func processDataFile(ctx context.Context, db *sql.DB, path string,
	stateData map[int]string) error {

	// open zipfile
	dataZip, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dataZip.Close()

	// get district data from zipfile
	districtData, err := getDistrictDataFromZip(dataZip)
	if err != nil {
		return err
	}
	defer districtData.Close()

	// update DB
	err = updateDatabase(ctx, db, districtData, stateData)
	if err != nil {
		return err
	}

	return nil
}

// ProcessCvap processes the CVAP data
func ProcessCvap(ctx context.Context, db *sql.DB, dataDirPath string) error {
	// load state data
	stateData, err := LoadStateData(dataDirPath)
	if err != nil {
		return err
	}

	// process CVAP files
	dataPath := path.Join(dataDirPath, "CVAP_2013-2017_ACS_csv_files.zip")
	log.Printf("Processing %v", dataPath)
	if err = processDataFile(ctx, db, dataPath, stateData); err != nil {
		return err
	}

	return nil
}
