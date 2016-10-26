package main

import "regexp"

type cmsPublicationEvent struct {
	ContentURI   string
	UUID         string
	Payload      interface{}
	LastModified string
}

func (e cmsPublicationEvent) matches(whiteList *regexp.Regexp) bool {
	return whiteList.MatchString(e.ContentURI)
}

func (e cmsPublicationEvent) hasEmptyPayload() bool {
	switch v := e.Payload.(type) {
	case nil:
		return true
	case string:
		if len(v) == 0 {
			return true
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return true
		}
	}
	return false
}
