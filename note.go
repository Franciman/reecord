package main

import (
    "time"
    "fmt"
    "strings"
)

// A single note taken
type Note struct {
    // This value is populated only when
    // we retrieve notes from the database
    // otherwise it's safe to leave it to its default value
    // For this value to be valid it must be > 0
    NoteID uint64

    Title string
    Link string
    Details string

    Author string

    // Date in which it was recorded
    Date time.Time
}

func (n *Note) PrettyPrintDate() string {
    year, month, day := n.Date.Date()
    return fmt.Sprintf("%d - %s - %d", day, month.String(), year)
}

func (n *Note) RenderLink() string {
    // Add URL schema if it's missing
    if !strings.HasPrefix(n.Link, "http") {
        return fmt.Sprintf("https://%s", n.Link)
    }
    return n.Link
}
