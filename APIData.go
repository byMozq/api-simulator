package main

type APIData struct {
	Method   string      `json:"method"`
	URL      string      `json:"url"`
	Request  APIRequest  `json:"request"`
	Response APIResponse `json:"response"`
}

type APIRequest struct {
	Headers map[string]string `json:"headers"`
	Body    APIBody           `json:"body"`
}

type APIResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       APIBody           `json:"body"`
}

type APIBody struct {
	Body string `json:"body"`
}
