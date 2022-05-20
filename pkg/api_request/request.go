package api_request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	GET    = "GET"
	POST   = "POST"
	PATCH  = "PATCH"
	PUT    = "PUT"
	DELETE = "DELETE"
)

func SendRequest(method string, url string, XAuthToken string, payloadMap interface{}, result interface{}) (response *http.Response, err error) {
	client := http.Client{}
	var request *http.Request
	if payloadMap != nil {
		payload, err := json.Marshal(payloadMap)
		if err != nil {
			return response, err
		}
		payloadReader := bytes.NewReader(payload)
		request, err = http.NewRequest(method, url, payloadReader)

	} else {
		request, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return
	}
	request.Header.Add("X-Auth-Token", XAuthToken)
	request.Header.Add("X-Subject-Token", XAuthToken)
	response, err = client.Do(request)
	if err != nil {
		return
	}
	if result != nil {
		responseBytes, err := ioutil.ReadAll(response.Body)
		if err = json.Unmarshal(responseBytes, result); err != nil {
			return response, err
		}
	}
	return
}
