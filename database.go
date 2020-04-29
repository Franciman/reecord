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
        link TEXT NOT NULL,
        details TEXT NOT NULL,
        author TEXT NOT NULL,
        date TEXT NOT NULL
        );

        CREATE TABLE IF NOT EXISTS users(
            username STRING PRIMARY KEY,
            password STRING
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

// If the note as a NoteID set, it will be ignored,
// because it is the database's responsibility to generate it
func (d Database) AddNote(note Note) error {
    tx, err := d.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("INSERT INTO notes(title, link, details, author, date) VALUES(?, ?, ?, ?, ?);")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(note.Title, note.Link, note.Details, note.Author, serializeDate(note.Date))
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (d Database) UpdateNote(note Note) error {
    tx, err := d.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("UPDATE notes SET title = ?, link = ?, details = ?, author = ?, date = ? WHERE id = ?;")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(note.Title, note.Link, note.Details, note.Author, serializeDate(note.Date), note.NoteID)
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

func (d Database) GetNote(noteID uint64) (*Note, error) {
    stmt, err := d.db.Prepare("SELECT id, title, link, details, author, date FROM notes WHERE id = ?")
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    rows, err := stmt.Query(noteID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var note *Note = nil
    for rows.Next() {
        var rawDate string
        note = new(Note)
        err = rows.Scan(&note.NoteID, &note.Title, &note.Link, &note.Details, &note.Author, &rawDate)
        if err != nil {
            return nil, err
        }
        note.Date, err = deserializeDate(rawDate)
        if err != nil {
            return nil, fmt.Errorf("Invalid date format: %v", err)
        }
        return note, nil
    }
    err = rows.Err()
    if err != nil {
        return nil, err
    }
    // If we are here, no note with the given id has been found
    return nil, nil

}

func (d Database) GetNotes() ([]Note, error) {
    rows, err := d.db.Query("SELECT id, title, link, details, author, date FROM notes")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    notes := make([]Note, 0)

    for rows.Next() {
        var note Note
        var rawDate string
    	err = rows.Scan(&note.NoteID, &note.Title, &note.Link, &note.Details, &note.Author, &rawDate)
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

func (d Database) CheckLogin(username, password string) (bool, error) {
    stmt, err := d.db.Prepare("SELECT password FROM users WHERE username = ?")
    if err != nil {
        return false, err
    }
    defer stmt.Close()

    rows, err := stmt.Query(username)
    if err != nil {
        return false, err
    }
    defer rows.Close()

    loginSuccess := false

    // Loop until there are users and until the password is wrong
    for rows.Next() && !loginSuccess {
        var pass string
        err = rows.Scan(&pass)
        if err != nil {
            return false, err
        }
        if pass == password {
            // Successful login
            loginSuccess = true
        }
    }

    err = rows.Err()
    if err != nil {
        return false, err
    }

    // If we are here, no error happened,
    // return whether the login was successful
    return loginSuccess, nil
}

func (d Database) HasUser(username string) (bool, error) {
    stmt, err := d.db.Prepare("SELECT username FROM users WHERE username = ?")
    if err != nil {
        return false, err
    }
    defer stmt.Close()

    rows, err := stmt.Query(username)
    if err != nil {
        return false, err
    }
    defer rows.Close()

    hasUser := false

    // This simply means check that there is at least one row in the result
    for rows.Next() && !hasUser {
        hasUser = true
    }

    err = rows.Err()
    if err != nil {
        return false, err
    }

    return hasUser, nil
}

// Returns false if the username is already taken
func (d Database) AddUser(username, password string) (bool, error) {
    // First check that the username has not already been taken
    usernameTaken, err := d.HasUser(username)
    if err != nil {
        return false, err
    }
    if usernameTaken {
        return false, nil
    }

    tx, err := d.db.Begin()
    if err != nil {
        return false, err
    }

    stmt, err := tx.Prepare("INSERT INTO users(username, password) VALUES(?, ?);")
    if err != nil {
        return false, err
    }
    defer stmt.Close()

    _, err = stmt.Exec(username, password)
    if err != nil {
        return false, err
    }

    if err := tx.Commit(); err != nil {
        return false, err
    }

    return true, nil
}

func (d Database) RemoveUser(username string) error {
    tx, err := d.db.Begin()
    if err != nil {
        return nil
    }

    stmt, err := tx.Prepare("DELETE FROM users WHERE username = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(username)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (d Database) ChangePassword(username, password string) error {
    tx, err := d.db.Begin()
    if err != nil {
        return err
    }

    stmt, err := tx.Prepare("UPDATE users SET password = ? WHERE username = ?")
    if err != nil {
        return err
    }
    defer stmt.Close()

    _, err = stmt.Exec(password, username)
    if err != nil {
        return err
    }

    return tx.Commit()
}
