package main

import (
	"context"
	"database/sql"
	"fmt"
)

type stateIrregularity struct {
	dbTable string
}

func (irreg *stateIrregularity) hasState(ctx context.Context, stateAbbr string,
	congress int) (bool, error) {

	sql := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE state = $1 AND congress_nbr = $2",
		irreg.dbTable)
	rows, err := gDb.QueryContext(ctx, sql, stateAbbr, congress)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var count int
	rows.Next()
	if err = rows.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

var gStateIrregularities = []stateIrregularity{
	stateIrregularity{dbTable: "state_with_atlarge_and_nonatlarge_districts"},
	stateIrregularity{dbTable: "state_with_overlapping_terms"},
	stateIrregularity{dbTable: "state_with_unknown_district"},
}

func getStateIrregularities(ctx context.Context, stateAbbr string,
	congress int) ([]string, error) {

	var irregularities []string
	for _, irregularity := range gStateIrregularities {
		hasIrreg, err := irregularity.hasState(ctx, stateAbbr, congress)
		if err != nil {
			return nil, err
		}
		if hasIrreg {
			irregularities = append(irregularities, irregularity.dbTable)
		}
	}
	return irregularities, nil
}

func getStates(ctx context.Context, congress int) ([]string, error) {
	sql := `SELECT DISTINCT state FROM house_district
	WHERE congress_nbr = $1`
	rows, err := gDb.QueryContext(ctx, sql, congress)
	if err != nil {
		return nil, err
	}

	var result []string
	for rows.Next() {
		var stateAbbr string
		if err = rows.Scan(&stateAbbr); err != nil {
			return nil, err
		}
		result = append(result, stateAbbr)
	}

	return result, nil
}

type districtInfo struct {
	rowID int
	nbr   int
}

func getDistricts(ctx context.Context, congress int) (map[string][]*districtInfo, error) {
	sql := `SELECT id, district, state FROM house_district WHERE congress_nbr = $1`
	rows, err := gDb.QueryContext(ctx, sql, congress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]*districtInfo)
	for rows.Next() {
		var di districtInfo
		var state string
		if err = rows.Scan(&di.rowID, &di.nbr, &state); err != nil {
			return nil, err
		}

		_, ok := result[state]
		if !ok {
			result[state] = nil
		}
		result[state] = append(result[state], &di)
	}

	return result, nil
}

func getDistrictFacts(ctx context.Context, districtID int) (districtFacts, error) {
	facts := make(districtFacts)
	var rows *sql.Rows
	var err error
	var sql string

	// get turnout
	sql = `SELECT turnout.num_votes, source.name
	FROM house_district_turnout AS turnout JOIN source
	ON (turnout.source_id = source.id)
	WHERE house_district_id = $1`
	rows, err = gDb.QueryContext(ctx, sql, districtID)
	if err != nil {
		goto done
	}
	if rows.Next() {
		var f fact
		if err = rows.Scan(&f.Value, &f.Source); err != nil {
			goto done
		}
		facts["turnout"] = &f
	}
	rows.Close()

	// get populations
	sql = `SELECT pop.type, pop.value, pop.margin_of_error, source.name
	FROM house_district_pop AS pop JOIN source
	ON (pop.source_id = source.id)
	WHERE house_district_id = $1`
	rows, err = gDb.QueryContext(ctx, sql, districtID)
	if err != nil {
		goto done
	}
	for rows.Next() {
		var f factWithMoe
		var fType string
		if err = rows.Scan(&fType, &f.Value, &f.MarginOfError, &f.Source); err != nil {
			goto done
		}
		facts[fType] = &f
	}
	rows.Close()

done:
	if rows != nil {
		rows.Close()
	}
	if err != nil {
		return nil, err
	}
	return facts, nil
}

func getDistrictID(ctx context.Context, congress int, state string,
	district int) (*int, error) {

	sql := `SELECT id FROM house_district
	WHERE congress_nbr = $1 AND state = $2 AND district = $3`
	rows, err := gDb.QueryContext(ctx, sql, congress, state, district)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}
	var id int
	if err = rows.Scan(&id); err != nil {
		return nil, err
	}
	return &id, nil
}
