package function

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

type Callable func(ignoreParams ...any) (resultInterfaces []any)

// Function struct stores the registered function
type Function struct {
	fn        any
	funcValue reflect.Value
	def       *Definition
	ignoreIdx []int
	fnInfo    *FuncInfo
}

// Definition provides the calling information of a function
type Definition struct {
	Name        string `json:"name"` // function name should be unique
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

func (d *Definition) String() string {
	jsonStr, _ := json.Marshal(d)
	return string(jsonStr)
}

func (f *Function) GetName() string {
	return f.def.Name
}

func (f *Function) GetDef() *Definition {
	return f.def
}

func (f *Function) GetIgnoreIdx() []int {
	return f.ignoreIdx
}

func (f *Function) GetFn() any {
	return f.fn
}

// CreateFunction creates a Function from a function
func CreateFunction(fn any, def Definition, ignoreIdx ...int) (*Function, error) {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()
	ftN := ft.NumIn() // total number of parameters

	// check ignoreParamsIdx: all of idx between 0 and ftN-1, remove duplicate idx, sort idx
	igMap := make(map[int]bool)
	for _, idx := range ignoreIdx {
		if idx >= ftN || idx < 0 {
			return nil, fmt.Errorf("invalid ignoreParamsIdx %d", idx)
		}
		// remove duplicate idx
		igMap[idx] = true
	}
	newIgnoreIdx := make([]int, 0, len(igMap))
	for k := range igMap {
		newIgnoreIdx = append(newIgnoreIdx, k)
	}
	sort.Ints(newIgnoreIdx)

	// check if the number of parameters in the def is the same as the function
	fcN := ftN - len(igMap)
	if def.Parameters == nil && fcN != 0 {
		return nil, fmt.Errorf("function %s does not match the number of parameters as the def (%d ignored)", def.Name, len(igMap))
	}
	//if def.Parameters != nil && fcN != len(def.Parameters.Properties) {
	//	return nil, fmt.Errorf("function %s does not match the number of parameters as the def (%d ignored)", def.Name, len(igMap))
	//}

	function := Function{
		fn:        fn,
		funcValue: fv,
		def:       &def,
		ignoreIdx: newIgnoreIdx,
	}
	return &function, nil
}

func (f *Function) GetCallable(p *Params) (Callable, error) {
	raw := p.IsRaw()
	var err error
	ignoreIdx := f.ignoreIdx
	ignoreIdxMap := make(map[int]bool)
	for _, idx := range ignoreIdx {
		ignoreIdxMap[idx] = true
	}
	ft := f.funcValue.Type()
	ftN := ft.NumIn()
	fcN := ftN - len(ignoreIdx)
	if p.Len() != fcN {
		return nil, fmt.Errorf("parameter count mismatch for function %s", f.GetName())
	}

	params := make([]reflect.Value, ftN)
	// i: idx of all params
	// j: idx of rawParams
	for i, j := 0, 0; i < ftN; i++ {
		if ignoreIdxMap[i] {
			continue
		}
		var paramValue reflect.Value
		if raw {
			rawParam := p.GetRaw(j)
			var paramType reflect.Type
			if i == ftN-1 && ft.IsVariadic() {
				// treat the last variadic parameter as a slice
				paramType = reflect.SliceOf(ft.In(i)).Elem() // Elem() is necessary
			} else {
				paramType = ft.In(i)
			}
			paramValueI := reflect.New(paramType).Interface()
			err = json.Unmarshal([]byte(rawParam), paramValueI)
			if err != nil {
				return nil, fmt.Errorf("invalid parameter: <%s> for function <%s>", rawParam, f.GetName())
			}
			paramValue = reflect.ValueOf(paramValueI).Elem()
		} else {
			paramValue = reflect.ValueOf(p.Get(j)) // todo test
		}

		// add to params
		if i == ftN-1 && ft.IsVariadic() {
			// for variadic parameter, expand the slice
			params = params[:len(params)-1] // remove the last element
			for k := 0; k < paramValue.Len(); k++ {
				params = append(params, paramValue.Index(k))
			}
		} else {
			params[i] = paramValue
		}
		j++
	}
	return func(ignoreParams ...any) (results []any) {
		if len(ignoreIdx) != len(ignoreParams) {
			panic(fmt.Errorf("function %s ignoreIdx and ignoreParams length mismatch, required %d ignoreParams, %d provided", f.GetName(), len(ignoreIdx), len(ignoreParams)))
		}
		for i, idx := range ignoreIdx {
			params[idx] = reflect.ValueOf(ignoreParams[i])
		}

		// Call the function and collect the results
		resultValues := f.funcValue.Call(params)
		results = make([]any, len(resultValues))
		for i, result := range resultValues {
			results[i] = result.Interface()
		}
		return results
	}, nil
}

func (f *Function) Call(params *Params, ignoreParams ...any) (resultInterfaces []any, err error) {
	callable, err := f.GetCallable(params)
	if err != nil {
		return nil, err
	}
	return callable(ignoreParams...), nil
}

func (f *Function) GetOrGenFuncInfo() (*FuncInfo, error) {
	if f.fnInfo == nil {
		fnInfo, err := GetFunctionDetails(f.fn)
		if err != nil {
			return nil, err
		}
		f.fnInfo = fnInfo
	}
	return f.fnInfo, nil
}

type Call struct {
	Name   string
	Params *Params
}

type Params struct {
	Params    []any
	RawParams []string // will be used when Params is nil, each RawParam is a json string
}

func (p *Params) IsRaw() bool {
	return p.Params == nil
}

func (p *Params) Len() int {
	if p.Params != nil {
		return len(p.Params)
	}
	return len(p.RawParams)
}

func (p *Params) GetRaw(i int) string {
	return p.RawParams[i]
}

func (p *Params) Get(i int) any {
	return p.Params[i]
}
