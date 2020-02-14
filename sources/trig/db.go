package trig

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

func Connect() (*DB, error) {
	connStr := "user=osmpoi dbname=osmpoi password=secret sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) CreateTable() error {
	q := `
	CREATE TABLE IF NOT EXISTS trigbeacons (
		name        			varchar NOT NULL,
		lat         			real NOT NULL,
		lon   					real NOT NULL,
		ele         			real NOT NULL,
		beacon_area_number      integer NOT NULL,
		beacon_number          	integer NOT NULL,
		description         	varchar,
		created_by 				varchar,
		osmid					bigint,
		PRIMARY KEY(beacon_area_number, beacon_number)
	);
`
	_, err := db.conn.Exec(q)
	return err
}

func (db *DB) QueryIncomplete(limit int) ([]*Trig, error) {
	rows, err := db.conn.Query(fmt.Sprintf("SELECT * from trigbeacons WHERE osmid = 0 LIMIT %d;", limit))
	if err != nil {
		return nil, err
	}
	trigs := make([]*Trig, 0, 4096)
	for rows.Next() {
		t := &Trig{Number: &BeaconNumber{}}
		err = rows.Scan(&t.Name, &t.Lat, &t.Lon, &t.Ele, &t.Number.Area, &t.Number.Number, &t.Description, &t.CreatedBy, &t.OSMID)
		if err != nil {
			return nil, err
		}
		trigs = append(trigs, t)
	}
	return trigs, nil
}

func (db *DB) UpdateOSMIDs(trigs []*Trig) error {
	for _, t := range trigs {
		err := db.UpdateOSMID(t)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) UpdateOSMID(t *Trig) error {
	_, err := db.conn.Exec(fmt.Sprintf("UPDATE trigbeacons SET osmid = %d WHERE beacon_area_number = %d AND beacon_number = %d;", t.OSMID, t.Number.Area, t.Number.Number))
	if err != nil {
		return fmt.Errorf("error inserting osmid %d for %s: %s", t.OSMID, t.Number, err)
	}
	return nil
}

func (db *DB) Import(trigs []*Trig) error {
	txn, err := db.conn.Begin()
	if err != nil {
		return err
	}
	stmt, err := txn.Prepare(pq.CopyIn("trigbeacons", "name", "lat", "lon", "ele", "beacon_area_number", "beacon_number", "description", "created_by", "osmid"))
	if err != nil {
		return err
	}
	dups := make(map[string]struct{}, len(trigs))
	for _, t := range trigs {
		if _, ok := dups[t.Number.String()]; ok {
			return fmt.Errorf("error inserting %#v: already exists", t)
		}
		dups[t.Number.String()] = struct{}{}
		_, err = stmt.Exec(t.Name, t.Lat, t.Lon, t.Ele, t.Number.Area, t.Number.Number, t.Description, t.CreatedBy, t.OSMID)
		if err != nil {
			return fmt.Errorf("error inserting %#v: %s", t, err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return txn.Commit()
}
