package main

import (
    "net/http"
    "html/template"
    "github.com/gin-gonic/gin"
    "log"
    "time"
    "strconv"
)

func serveIndexPage(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        notes, err := db.GetNotes()
        if err != nil {
            log.Printf("Error while retrieving notes from db: %v", err)
            c.AbortWithStatus(http.StatusInternalServerError)
        } else {
            c.HTML(http.StatusOK, "index.html", notes)
        }
    }
}

func addNote(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {

        title := c.PostForm("title")
        details := c.PostForm("details")

        if title == "" {
            c.AbortWithStatus(http.StatusBadRequest)
            return
        }

        noteDate := time.Now()

        if err := db.AddNote(title, details, noteDate); err != nil {
            log.Println("Error while adding note to db: ", err)
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }

        c.Redirect(http.StatusSeeOther, "/")
    }
}

func deleteNote(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        rawNoteID := c.PostForm("note_id")
        noteID, err := strconv.ParseUint(rawNoteID, 10, 64)
        if err != nil {
            log.Println("Error while parsing noteID: ", err)
            c.AbortWithStatus(http.StatusBadRequest)
            return
        }
        if err := db.RemoveNote(noteID); err != nil {
            log.Println("Error while removing note from db: ", err)
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }
        c.Redirect(http.StatusSeeOther, "/")
    }
}

func getFile(path string) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.File(path)
    }
}

func StartServer(addr string, tmpl *template.Template, db Database) error {

    r := gin.Default()
    r.SetHTMLTemplate(tmpl)

    r.GET("/style.css", getFile("assets/style.css"))
    r.GET("/", serveIndexPage(db))

    r.POST("/add_note", addNote(db))
    r.POST("/delete_note", deleteNote(db))

    return r.Run(addr)
}
