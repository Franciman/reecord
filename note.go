package main

import (
    "time"
    "fmt"
)

// A single note taken
type Note struct {
    // This value is populated only when
    // we retrieve notes from the database
    // otherwise it's safe to leave it to its default value
    // For this value to be valid it must be > 0
    NoteID uint64

    Title string
    Details string

    // Date in which it was recorded
    Date time.Time
}

func (n *Note) PrettyPrintDate() string {
    year, month, day := n.Date.Date()
    return fmt.Sprintf("%d - %s - %d", day, month.String(), year)
}
