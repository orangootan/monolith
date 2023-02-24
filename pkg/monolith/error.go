package monolith

import (
	"encoding/gob"
)

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func NewError(message string) Error {
	return Error{
		Message: message,
	}
}

var UnregisteredTypeError = NewError("unregistered type request received")
var MethodNotFoundError = NewError("method not found")
var RequestNotFoundError = NewError("request not found")
var ServiceNotFoundError = NewError("service not found")
var ProxyTypeNotFoundError = NewError("proxy type not found")

func init() {
	gob.Register(NewError(""))
}
