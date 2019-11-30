package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"os/signal"
)

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

func run(dataDirPath string) error {
	// connect to DB
	connStr := "host=db user=postgres password=pw dbname=house sslmode=disable connect_timeout=10"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	// launch signal listener
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	handleSignals(stop)

	// load data
	log.Printf("Processing congress data")
	if err = UpdateCongresses(ctx, db, dataDirPath); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "COMMIT")
	if err != nil {
		return err
	}
	log.Printf("Processing CVAP data")
	if err = ProcessCvap(ctx, db, dataDirPath); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "COMMIT")
	if err != nil {
		return err
	}
	log.Printf("Processing turnout data")
	if err = UpdateTurnout(ctx, db, dataDirPath); err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, "COMMIT")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// parse args
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("usage: loaddata PATH_TO_DATA_DIR")
		os.Exit(1)
	}
	dataDirPath := flag.Arg(0)

	if err := run(dataDirPath); err != nil {
		log.Fatal(err)
	}
}
