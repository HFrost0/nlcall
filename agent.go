package nlcall

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HFrost0/nlcall/function"
	"os"
	"reflect"
	"runtime"
)

type Agent struct {
	resolver Resolver
	definer  Definer
	funcMap  map[string]*function.Function
	funcKeys []string
}

func NewAgent(resolver Resolver, definer Definer) *Agent {
	return &Agent{
		resolver: resolver,
		definer:  definer,
		funcMap:  make(map[string]*function.Function),
	}
}

// AssignFunc assigns the user input to the corresponding function.Function
func (a *Agent) AssignFunc(ctx context.Context, userInput string) (f *function.Function, rawParams []string, err error) {
	if userInput == "" {
		return nil, nil, EmptyUserInputErr
	}
	call, err := a.resolver.Resolve(ctx, userInput)
	if err != nil {
		return nil, nil, err
	}
	f, err = a.GetFunc(call.Name)
	if err != nil {
		return nil, nil, err
	}
	return f, rawParams, nil
}

func (a *Agent) AssignCallable(ctx context.Context, userInput string) (callable function.Callable, err error) {
	if userInput == "" {
		return nil, EmptyUserInputErr
	}
	call, err := a.resolver.Resolve(ctx, userInput)
	if err != nil {
		return nil, err
	}
	f, err := a.GetFunc(call.Name)
	if err != nil {
		return nil, err
	}
	callable, err = f.GetCallable(call.Params)
	if err != nil {
		return nil, err
	}
	return callable, nil
}

// RegisterFunc registers a function.Function to be called
func (a *Agent) RegisterFunc(f *function.Function) error {
	name := f.GetName()
	// check if the function name is already registered
	if _, ok := a.funcMap[name]; ok {
		return fmt.Errorf("function %s already exists", name)
	}
	a.funcMap[name] = f
	a.funcKeys = append(a.funcKeys, name)
	if ok := a.resolver.AddFunc(f); !ok {
		return fmt.Errorf("failed to add function %s to resolver", name)
	}
	return nil
}

type RegisterOption func(*RegisterOpts)

type RegisterOpts struct {
	LoadDefDir string // the dir to load function definition
	SaveDefDir string // the dir to save function definition
	Overwrite  bool   // whether to overwrite the existing definition
}

func WithLoadDefDir(path string) RegisterOption {
	return func(o *RegisterOpts) {
		o.LoadDefDir = path
	}
}

func WithSaveDefDir(path string) RegisterOption {
	return func(o *RegisterOpts) {
		o.SaveDefDir = path
	}
}

func WithOverwrite(overwrite bool) RegisterOption {
	return func(o *RegisterOpts) {
		o.Overwrite = overwrite
	}
}

func buildRegisterOpts(opts ...RegisterOption) RegisterOpts {
	var registerOpts RegisterOpts
	for _, opt := range opts {
		opt(&registerOpts)
	}
	return registerOpts
}

// RegisterFn registers a golang function can be called
// the golang fn definition can be generated by the Definer according to options
// ignore idx in this case will not be considered
func (a *Agent) RegisterFn(ctx context.Context, fn any, opts ...RegisterOption) (*function.Function, error) {
	var def *function.Definition
	var err error
	registerOpts := buildRegisterOpts(opts...)
	if registerOpts.LoadDefDir != "" {
		def, err = loadDefFromDisk(registerOpts.LoadDefDir, getFnName(fn))
		if os.IsNotExist(err) {
			// create def by Definer
			def, err = a.definer.Define(ctx, fn)
		}
	} else {
		// create def by Definer
		def, err = a.definer.Define(ctx, fn)
	}
	if err != nil {
		return nil, err
	}

	f, err := function.CreateFunction(fn, *def)
	if err != nil {
		return nil, err
	}
	if err = a.RegisterFunc(f); err != nil {
		return nil, err
	}
	if registerOpts.SaveDefDir != "" {
		err = saveDefToDisk(registerOpts.SaveDefDir, getFnName(fn), def, registerOpts.Overwrite)
		if err != nil {
			return nil, err
		}
	}
	return f, nil
}

// GetFunc looks up the registered function by name
func (a *Agent) GetFunc(funcName string) (*function.Function, error) {
	// Look up the function
	f, ok := a.funcMap[funcName]
	if !ok {
		return nil, fmt.Errorf("function %s does not exist", funcName)
	}
	return f, nil
}

const defFileSuffix = ".lcdef.json"

func loadDefFromDisk(dir string, fnName string) (*function.Definition, error) {
	path := fmt.Sprintf("%s/%s%s", dir, fnName, defFileSuffix)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	def := new(function.Definition)
	err = json.Unmarshal(bytes, def)
	if err != nil {
		return nil, err
	}
	return def, nil
}

func saveDefToDisk(dir string, fnName string, def *function.Definition, overwrite bool) error {
	path := fmt.Sprintf("%s/%s%s", dir, fnName, defFileSuffix)
	bytes, err := json.Marshal(def)
	if err != nil {
		return err
	}
	if !overwrite {
		if _, err = os.Stat(path); err == nil {
			return nil
		}
		err = os.WriteFile(path, bytes, 0644)
	} else {
		err = os.WriteFile(path, bytes, 0644)
	}
	if err != nil {
		return err
	}
	return nil
}

func getFnName(fn any) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
