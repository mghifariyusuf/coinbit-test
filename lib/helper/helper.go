package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var defaultErrorResp []byte

func GetBufBody(r *http.Request) (buf bytes.Buffer, err error) {
	_, err = buf.ReadFrom(r.Body)
	return buf, err
}

func ResponseObject(w http.ResponseWriter, object interface{}) {
	status := http.StatusOK

	jsonResp, err := json.Marshal(object)
	if err != nil {
		jsonResp = defaultErrorResp
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(jsonResp)
}

func ResponseError(r *http.Request, w http.ResponseWriter, httpStatusCode int, err error) {
	resp := map[string]interface{}{
		"type":    http.StatusText(httpStatusCode),
		"message": err.Error(),
	}

	jsonResp, errMarshal := json.Marshal(resp)
	if errMarshal != nil {
		jsonResp = defaultErrorResp
		httpStatusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	_, _ = w.Write(jsonResp)
}
