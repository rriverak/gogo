package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Logger is the API Logger Instance
var Logger *logrus.Logger

//WriteJSON to Response
func WriteJSON(w http.ResponseWriter, v interface{}) {
	data, _ := json.Marshal(v)
	WriteContentType(w, "application/json")
	WriteData(w, data)
}

//WriteData to Response with Status OK
func WriteData(w http.ResponseWriter, data []byte) {
	WriteStatusOK(w)
	fmt.Fprintf(w, string(data))
}

//WriteError to Response as JSON with Status 500
func WriteError(w http.ResponseWriter, err error) {
	WriteStatusError(w)
	WriteContentType(w, "application/json")
	data, _ := json.Marshal(map[string]string{"error": err.Error()})
	fmt.Fprintf(w, string(data))
}

//WriteContentType to Response Header
func WriteContentType(w http.ResponseWriter, cType string) {
	w.Header().Set("Content-Type", cType)
}

//WriteStatusOK to Response Header
func WriteStatusOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

//WriteStatusNotFound to Response Header
func WriteStatusNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

//WriteStatusError to Response Header
func WriteStatusError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
