package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//DecodeBody a HTTP Body
func DecodeBody(r *http.Request, v interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(&v)
}

//WriteRedirect to Response Header
func WriteRedirect(w http.ResponseWriter, path string) {
	w.Header().Set("Location", path)
	w.WriteHeader(http.StatusSeeOther)
}

//WriteResultOrError to Response
func WriteResultOrError(w http.ResponseWriter, v interface{}, err error) {
	if err != nil {
		WriteError(w, err)
	} else {
		WriteJSON(w, v)
	}
}

//WriteJSON to Response
func WriteJSON(w http.ResponseWriter, v interface{}) {
	data, _ := json.Marshal(v)
	WriteContentType(w, "application/json")
	WriteData(w, data)
}

//WriteText to Response with Status OK
func WriteText(w http.ResponseWriter, text string) {
	WriteStatusOK(w)
	fmt.Fprintf(w, text)
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

//WriteStatusForbidden to Response Header
func WriteStatusForbidden(w http.ResponseWriter) {
	w.WriteHeader(http.StatusForbidden)
}

//WriteStatusNotFound to Response Header
func WriteStatusNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

//WriteStatusConfict to Response Header
func WriteStatusConfict(w http.ResponseWriter) {
	w.WriteHeader(http.StatusConflict)
}

//WriteStatusError to Response Header
func WriteStatusError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}
