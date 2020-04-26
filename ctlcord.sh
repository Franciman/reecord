#!/bin/bash

ADDRESS="localhost:9999"

case "$1" in
    "add-user")
        printf "Username: "
        read -r username
        printf "Password: "
        read -r -s password
        echo
        curl "$ADDRESS/add_user" -X POST -d $(printf 'username=%s&password=%s' "$username" "$password")
        echo
        ;;
    "remove-user")
        printf "Username: "
        read -r username
        curl "$ADDRESS/remove_user" -X POST -d $(printf 'username=%s' "$username")
        echo
esac

unset username password
