package main

import (
    "net/http"
    "io"
    //"encoding/json"
)

func main() {
    http.Handle("/", http.FileServer(http.Dir("../client/public")))
    http.HandleFunc("/function_handler", functionHandler)
    http.HandleFunc("/endpoints/server_name", dataEndpoint)
    http.ListenAndServe(":8080", nil)
}

func functionHandler(res http.ResponseWriter, req *http.Request) {
    io.WriteString(res, `
<!DOCTYPE html>
<html>
  <head></head>
  <body>
      This was served from a function handler.
  </body>
</html>`)
}

func dataEndpoint(res http.ResponseWriter, req *http.Request) {
    res.Header().Set("Content-Type", "application/json")
    res.WriteHeader(http.StatusOK)
    res.Write([]byte(`{"Name":"Joe Blogs"}`))
}

/*
Must create react files in sub directory: ./react
This can be helped with a compiler like babel
*/
