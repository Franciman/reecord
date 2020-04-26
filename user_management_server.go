package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "log"
)

func addUser(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        username := c.PostForm("username")
        password := c.PostForm("password")

        if username == "" || password == "" {
            c.JSON(http.StatusBadRequest, gin.H {
                "error": "Username and Password must be non empty.",
            })
            return
        }

        added, err := db.AddUser(username, password)
        if err != nil {
            log.Println("Error while adding user: ", err)
            c.JSON(http.StatusInternalServerError, gin.H {
                "error": "Internal server error, see logs.",
            })
            return
        }

        // This means that there was a duplicate user
        if !added {
            c.JSON(http.StatusBadRequest, gin.H {
                "error": "Username already taken.",
            })
            return
        }

        c.JSON(http.StatusOK, gin.H {
            "result": "User successfully added.",
        })
    }
}

func removeUser(db Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        // TODO: Probably this functionality should require the login info
        // but probably not
        username := c.PostForm("username")

        if username == "" {
            c.JSON(http.StatusBadRequest, gin.H {
                "error": "Username must be non empty.",
            })
            return
        }

        err := db.RemoveUser(username)
        if err != nil {
            log.Println("Error while removing user: ", err)
            c.JSON(http.StatusInternalServerError, gin.H {
                "error": "Internal server error, see logs.",
            })
            return
        }

        c.JSON(http.StatusOK, gin.H {
            "result": "User successfully removed.",
        })
    }
}

func StartUserManagementServer(addr string, db Database) error {
    r := gin.Default()

    r.POST("/add_user", addUser(db))
    r.POST("/remove_user", removeUser(db))

    return r.Run(addr)
}
