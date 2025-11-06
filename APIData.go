package main

import (
	"github.com/hashicorp/go-memdb"
)

type APIData struct {
	ID       string      // uuid
	Method   string      `json:"method"`
	URL      string      `json:"url"`
	Request  APIRequest  `json:"request"`
	Response APIResponse `json:"response"`
}

type APIRequest struct {
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type APIResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// GetAPIDataSchema creates the memdb schema for APIData
func GetAPIDataSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"apidata": {
				Name: "apidata",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UUIDFieldIndex{Field: "ID"},
					},
					"method_url": {
						Name:         "method_url",
						Unique:       false,
						AllowMissing: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Method"},
								&memdb.StringFieldIndex{Field: "URL"},
							},
							AllowMissing: true,
						},
					},
				},
			},
		},
	}
}
