package main

import (
    "fmt"
    "time"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)
const HISTORY_TABLE_NAME = "garage_history"
const COLUMN_GARAGE_ID = "garage_id"
const COLUMN_TIMESTAMP = "timestamp"
const COLUMN_STATUS_CHANGE = "status"

/* Sqlite code */
func CreateHistoryTableIfNeeded(db *sql.DB) error {
    // Build the sql statement for creating the table
    sqlStmt := "create table if not exists " + HISTORY_TABLE_NAME + " (id integer not null primary key, " +
    COLUMN_GARAGE_ID + " text not null, " + COLUMN_TIMESTAMP + " timestamp default CURRENT_TIMESTAMP not null, " +
    COLUMN_STATUS_CHANGE + " text check( "+COLUMN_STATUS_CHANGE+" in ('open','closed') ) not null);"

    log.Debugf("Create table statement:\n'%s'", sqlStmt);

    // Execute the statement
    _,err := db.Exec(sqlStmt)

    return err
}

func WriteHistoryEvent(db *sql.DB, update UpdateMessage) error {
    // Begin the transaction
    tx, err := db.Begin()
    if err != nil { return err }

    // Prepare the statement
    strStmt := fmt.Sprintf("insert into %s (%s, %s, %s) values (?, ?, ?)",
                HISTORY_TABLE_NAME, COLUMN_GARAGE_ID, COLUMN_STATUS_CHANGE, COLUMN_TIMESTAMP)
    log.Debugf("Preparing history string statement: '%s'", strStmt)

    stmt, err := tx.Prepare(strStmt)
    if err != nil { return err }

    defer stmt.Close()

    // execute the statement with the correct values
    _, err = stmt.Exec(update.DoorName, update.Status, update.LastChanged)
    if err != nil { return err }

    err = tx.Commit()
    if err != nil { return err }

    return nil
}

func QueryHistoryEvents(db *sql.DB, request HistoryRequestMessage) ([]HistoryResponse, error) {
    // Prepare the select statement
    var strStmt string

    strStmt = fmt.Sprintf("SELECT %s,%s,%s FROM %s WHERE %s BETWEEN ? AND ?",
            COLUMN_GARAGE_ID, COLUMN_STATUS_CHANGE, COLUMN_TIMESTAMP, HISTORY_TABLE_NAME, COLUMN_TIMESTAMP)

    // Only get the most recent value
    if request.CurrentValue == true {
        strStmt = fmt.Sprintf("SELECT %s,%s,%s FROM %s ORDER BY %s DESC LIMIT 1",
                COLUMN_GARAGE_ID, COLUMN_STATUS_CHANGE, COLUMN_TIMESTAMP, HISTORY_TABLE_NAME, COLUMN_TIMESTAMP)
    }

    log.Debugf("Preparing history select statement: '%s'", strStmt)

    stmt, err := db.Prepare(strStmt)
    if err != nil { return nil,err }

    defer stmt.Close()

    // Execute the statement
    var rows *sql.Rows

    // Only need the start and end times when we're not requesting the current value
    if request.CurrentValue == false {
        rows, err = stmt.Query(request.StartUnixTime, request.EndUnixTime)
    }else {
        rows, err = stmt.Query()
    }

    if err != nil { return nil,err }

    // Store records in a slice
    var events = make([]HistoryResponse, 0)

    for rows.Next() {
        var garageID string
        var status string
        var timestamp time.Time

        err = rows.Scan(&garageID, &status, &timestamp)
        if err != nil { return nil,err }

        response := HistoryResponse{ DoorName: garageID, Timestamp: timestamp, Status: status }

        events = append(events, response)
    }
    if rows.Err() != nil { return nil,err }

    return events,nil
}
