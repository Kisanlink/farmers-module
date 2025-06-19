package handlers

import "net/http"

// semanticStatus maps short keys to the code we agreed on in the review
var semanticStatus = struct {
	Validation int
	Dependency int
	NotFound   int
	Conflict   int
	Permission int
}{
	Validation: http.StatusUnprocessableEntity, // 422
	Dependency: http.StatusBadGateway,          // 502
	NotFound:   http.StatusNotFound,            // 404
	Conflict:   http.StatusConflict,            // 409
	Permission: http.StatusForbidden,           // 403
}
