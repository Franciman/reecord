package main

import (
    "net/http"
    "html/template"
    "log"
    "time"
    "strconv"
)

type HandleFunc = func(w http.ResponseWriter, req *http.Request)

func addHandler(method string, pattern string, handler HandleFunc) {
    http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
        if req.Method == method {
            handler(w, req)
        } else {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
}

func addGetHandler(pattern string, handler HandleFunc) {
    addHandler(http.MethodGet, pattern, handler)
}

func addPostHandler(pattern string, handler HandleFunc) {
    addHandler(http.MethodPost, pattern, handler)
}

func serveIndexPage(tmpl *template.Template, db Database) HandleFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        notes, err := db.GetNotes()
        if err != nil {
            log.Printf("Error while retrieving notes from db: %v", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        } else {
            tmpl.ExecuteTemplate(w, "index.html", notes)
        }
    }
}

func addNote(db Database) HandleFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := r.ParseForm(); err != nil {
            log.Println("Error while parsing form: ", err)
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }

        title := r.PostForm.Get("title")
        details := r.PostForm.Get("details")

        if title == "" {
            http.Error(w, "Invalid empty title", http.StatusBadRequest)
            return
        }

        noteDate := time.Now()

        if err := db.AddNote(title, details, noteDate); err != nil {
            log.Println("Error while adding note to db: ", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func deleteNote(db Database) HandleFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := r.ParseForm(); err != nil {
            log.Println("Error while parsing form: ", err)
            http.Error(w, "Bad Request", http.StatusBadRequest)
            return
        }

        rawNoteID := r.PostForm.Get("note_id")
        noteID, err := strconv.ParseUint(rawNoteID, 10, 64)
        if err != nil {
            log.Println("Error while parsing noteID: ", err)
            http.Error(w, "Invalid note id", http.StatusBadRequest)
            return
        }
        if err := db.RemoveNote(noteID); err != nil {
            log.Println("Error while removing note from db: ", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}

func StartServer(addr string, tmpl *template.Template, db Database) error {

    addGetHandler("/style.css", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./style.css")
    })
    addGetHandler("/", serveIndexPage(tmpl, db))
    addPostHandler("/add_note", addNote(db))
    addPostHandler("/delete_note", deleteNote(db))

    return http.ListenAndServe(addr, nil)
}
