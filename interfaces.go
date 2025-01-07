package nlcall

import (
	"context"
	"github.com/HFrost0/nlcall/function"
)

// Resolver resolves the user input to a function name and its parameters
type Resolver interface {
	AddFunc(def *function.Function) bool
	Resolve(ctx context.Context, userInput string) (call *function.Call, err error)
}

// Definer defines a function from golang func
type Definer interface {
	Define(ctx context.Context, fn any) (*function.Definition, error)
}
