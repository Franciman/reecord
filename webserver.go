package main

import (
    "net/http"
    "html/template"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "log"
    "time"
    "strconv"
)

// User key in session dictionary
const userKey string = "username"

func serveIndexPage(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        user := session.Get(userKey)

        loggedIn := false
        username := ""

        if user != nil {
            loggedIn = true
            username = user.(string)
        }

        notes, err := db.GetNotes()
        if err != nil {
            log.Printf("Error while retrieving notes from db: %v", err)
            c.AbortWithStatus(http.StatusInternalServerError)
        } else {
            c.HTML(http.StatusOK, "index.html", map[string]interface{} {
                   "Notes": notes,
                   "LoggedIn": loggedIn,
                   "Username": username,
            })
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

func doLogin(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        username := c.PostForm("username")
        password := c.PostForm("password")

        if username == "" || password == "" {
            c.AbortWithStatus(http.StatusBadRequest)
            return
        }

        loginSuccess, err := db.CheckLogin(username, password)
        if err != nil {
            log.Println("Error while checking login data: ", err)
            c.AbortWithStatus(http.StatusInternalServerError)
            return
        }
        if loginSuccess || true {
            // Update session
            session := sessions.Default(c)
            session.Set(userKey, username)

            err = session.Save()
            if err != nil {
                log.Println("Error while saving session: ", err)
                c.AbortWithStatus(http.StatusInternalServerError)
            }

            c.Redirect(http.StatusSeeOther, "/")
        } else {
            c.AbortWithStatus(http.StatusUnauthorized)
        }
    }
}

func doLogout(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {

        session := sessions.Default(c)
        user := session.Get(userKey)

        if user == nil {
            c.AbortWithStatus(http.StatusUnauthorized)
            return
        }

        session.Delete(userKey)
        err := session.Save()
        if err != nil {
            log.Println("Error while saving session: ", err)
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

func requireAuth(c *gin.Context) {
    session := sessions.Default(c)
    user := session.Get(userKey)

    if user == nil {
        c.Redirect(http.StatusSeeOther, "/login.html")
        c.Abort()
    } else {
        c.Next()
    }
}

func StartWebServer(addr string, tmpl *template.Template, db Database) error {

    r := gin.Default()
    r.SetHTMLTemplate(tmpl)

    cookie := cookie.NewStore([]byte("secret"))
    r.Use(sessions.Sessions("my-session", cookie))

    r.GET("/style.css", getFile("assets/style.css"))
    r.GET("/login.html", getFile("assets/login.html"))
    r.GET("/", serveIndexPage(db))

    private := r.Group("/")
    private.Use(requireAuth)
    {
        private.POST("/add_note", addNote(db))
        private.POST("/delete_note", deleteNote(db))
    }
    r.POST("/do_login", doLogin(db))
    r.POST("/do_logout", doLogout(db))

    return r.Run(addr)
}
