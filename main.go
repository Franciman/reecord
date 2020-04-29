package main

import (
    "log"
    "html/template"
)

func main() {
    tmpl, err := template.New("reecord").ParseFiles("assets/index.html", "assets/update_note.html")
    if err != nil {
        log.Fatal("Error while parsing html files ", err)
        return
    }

    db, err := OpenDatabase("./reecord.db")
    if err != nil {
        log.Fatal("Error while opening database: ", err)
        return
    }

    go func() {
        if err := StartUserManagementServer(":9999", db); err != nil {
            log.Println("UserManagementServer error: ", err)
        }
    }()

    if err := StartWebServer(":8888", tmpl, db); err != nil {
        log.Fatal("Webserver error: ", err)
    }
}
