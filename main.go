package main

import (
    "log"
    "html/template"
)

func main() {
    tmpl, err := template.New("reecord").ParseFiles("index.html")
    if err != nil {
        log.Fatal("Error while parsing html files ", err)
        return
    }

    db, err := OpenDatabase("./reecord.db")
    if err != nil {
        log.Fatal("Error while opening database: ", err)
        return
    }

    StartServer(":8888", tmpl, db)
}
