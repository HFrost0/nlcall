package nlcall

import "errors"

var (
	EmptyUserInputErr = errors.New("empty user input")
)

type FuncCreateErr struct {
	Msg string
}

type FuncCallErr struct {
	Msg string
}

type FuncStrParseErr struct {
	Msg string
}

func (e FuncCreateErr) Error() string {
	return e.Msg
}

func (e FuncCallErr) Error() string {
	return e.Msg
}

func (e FuncStrParseErr) Error() string {
	return e.Msg
}
