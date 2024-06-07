package main

import (
	"net/http"
	"real-time-message/handler"
)

func main() {
	http.HandleFunc("/", handler.InitWs)

	http.ListenAndServe(":8005", nil)
}
