package domain

import (
	"encoding/json"
	logs "github.com/certified-juniors/AtomHack/internal/logger"
	"io"
	"net/http"
)

type Response struct {
	Body interface{} `json:"body,omitempty"`
	Err  string      `json:"err,omitempty"`
}

func WriteError(w http.ResponseWriter, errString string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&Response{Err: errString})
}

func WriteResponse(w http.ResponseWriter, result map[string]interface{}, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&Response{Body: result})
}

func CloseAndAlert(body io.ReadCloser, packageName, funcName string) {
	err := body.Close()
	if err != nil {
		logs.LogError(logs.Logger, packageName, funcName, err, err.Error())
	}
}
