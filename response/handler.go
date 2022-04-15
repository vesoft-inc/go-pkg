package response

import "net/http"

type (
	Handler interface {
		Handle(w http.ResponseWriter, r *http.Request, data interface{}, err error)
		GetStatusBody(r *http.Request, data interface{}, err error) (httpStatus int, body interface{})
	}
)
