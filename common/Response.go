package common

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func BuildResponse(resp http.ResponseWriter, code int, msg string, data interface{}) {
	var (
		response Response
	)
	response.Code = code
	response.Msg = msg
	response.Data = data

	// 2, 序列化json
	resp.Header().Set("Context-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(resp)

	if err := encoder.Encode(response); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}
