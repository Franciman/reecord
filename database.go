package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
    "fmt"
)

type Database struct {
    db *sql.DB
}

func (d Database) setupSchema() error {
    stmt := `CREATE TABLE IF NOT EXISTS notes(
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        details TEXT NOT NULL,
        date TEXT NOT NULL
        );`

    _, err := d.db.Exec(stmt)
    return err

}


func OpenDatabase(filename string) (Database, error) {
    db, err := sql.Open("sqlite3", filename)
    if err != nil {
        return Database{}, err
    }

    d := Database { db }
    err = d.setupSchema()
    if err != nil {
        // Close the database!
        d.CloseDatabase()
        return Database{}, err
    }

    return d, nil
}

func (d Database) CloseDatabase() error {
    return d.db.Close()
}

// Helper functions to convert between ISO8601 text format and time.Time type
// We store dates as text in the database,
// but we want to use time.Time go datatype in code

var dateFormat string = "2006-01-02 15:04:05"

func serializeDate(date time.Time) string {
    return date.Format(dateFormat)
}

func deserializeDate(date string) (time.Time, error) {
    return time.Parse(dateFormat, date)
}

func (d Database) AddNote(title string, details string, date time.Time) error {
    tx, err := d.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("INSERT INTO notes(title, details, date) VALUES(?, ?, ?);")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(title, details, serializeDate(date))
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (d Database) RemoveNote(noteID uint64) error {
    tx, err := d.db.Begin()
    if err != nil {
        return nil
    }

    stmt, err := tx.Prepare("DELETE FROM notes WHERE id = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(noteID)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (d Database) GetNotes() ([]Note, error) {
    rows, err := d.db.Query("SELECT id, title, details, date FROM notes")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    notes := make([]Note, 0)

    for rows.Next() {
        var note Note
        var rawDate string
    	err = rows.Scan(&note.NoteID, &note.Title, &note.Details, &rawDate)
    	if err != nil {
    	    return nil, err
    	}
    	note.Date, err = deserializeDate(rawDate)
    	if err != nil {
    	    return nil, fmt.Errorf("Invalid date format: %v", err)
    	}
    	notes = append(notes, note)
    }
    err = rows.Err()
    if err != nil {
        return nil, err
    }

    return notes, nil
}
