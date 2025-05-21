package main

type request struct {
	url     string
	method  string
	headers map[string]string
	body    string

	response
}

type response struct {
	statusCode int
	headers    map[string]string
	body       string
}
