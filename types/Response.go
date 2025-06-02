package types

type Response struct {
	Error string "json:\"error\""
	Code  string "json:\"code\""
	Data  any    "json:\"data\""
}
