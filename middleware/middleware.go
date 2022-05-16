package middleware

import "net/http"

type Skipper func(*http.Request) bool

func DefaultSkipper(*http.Request) bool {
	return false
}
