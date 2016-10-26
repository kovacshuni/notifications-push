package main

type notification struct {
	APIURL           string `json:"apiUrl"`
	ID               string `json:"id"`
	Type             string `json:"type"`
	PublishReference string `json:"publishReference,omitempty"`
	LastModified     string `json:"lastModified,omitempty"`
}
