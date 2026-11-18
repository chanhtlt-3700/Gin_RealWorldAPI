package dto

// ErrorResponse là format lỗi chuẩn của RealWorld:
//
//	{"errors": {"body": ["can't be empty"]}}
type ErrorResponse struct {
	Errors map[string][]string `json:"errors"`
}

func NewErrorResponse(errs map[string][]string) ErrorResponse {
	return ErrorResponse{Errors: errs}
}
