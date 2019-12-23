package tuftsTurnout

import (
	"context"
	"database/sql"
	"log"
)

type stateSet map[string]bool

func (self stateSet) contains(state string) bool {
	_, ok := self[state]
	return ok
}

func (self stateSet) insert(state string) {
	self[state] = true
}

var gIrregStates map[int]stateSet

func loadIrregStates(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT state, congress_nbr FROM irregular_state")
	if err != nil {
		return err
	}
	defer rows.Close()

	gIrregStates = make(map[int]stateSet)
	for rows.Next() {
		var state string
		var congressNbr int
		if err = rows.Scan(&state, &congressNbr); err != nil {
			return err
		}

		states, ok := gIrregStates[congressNbr]
		if !ok {
			states = make(stateSet)
			gIrregStates[congressNbr] = states
		}
		states.insert(state)
	}

	return nil
}

func isIrregState(state string, congressNbr int) bool {
	if gIrregStates == nil {
		log.Fatal("Didn't load irregular states")
	}

	states, ok := gIrregStates[congressNbr]
	if !ok {
		return false
	}
	return states.contains(state)
}
