package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HFrost0/nlcall/function"
	"regexp"
	"strings"
)

var sysPromptTemplate = `there are some functions defined below:
'''
%s
'''

Your task is to choose a suitable function and output a formated calling string: '''<func_name>(<arg1>,<arg2>,...)'''
output by the rules:
1. output without any explanation.
2. there is no space between the arguments since you need to save the space.
3. parameters is null means you should not pass any arguments.
`
var funcRex = regexp.MustCompile(`(\w+)\((.*)\)`)

type Resolver struct {
	completionClient         CompletionClient
	completionWithToolClient CompletionWithToolClient
	sysPrompt                string
	sysPromptTemplate        string
	fnName2fn                map[string]*function.Function
	fnNames                  []string
}

func NewResolver(completionClient CompletionClient) *Resolver {
	r := &Resolver{
		completionClient:  completionClient,
		sysPromptTemplate: sysPromptTemplate,
		fnName2fn:         make(map[string]*function.Function),
	}
	if v, ok := completionClient.(CompletionWithToolClient); ok {
		r.completionWithToolClient = v
	}
	return r
}

func (r *Resolver) Resolve(ctx context.Context, userInput string) (call *function.Call, err error) {
	if r.completionWithToolClient != nil {
		return r.resolveByTool(ctx, userInput)
	}
	return r.resolveByPrompt(ctx, userInput)
}

// resolveByTool resolves the user input to a function call by trained function calling
func (r *Resolver) resolveByTool(ctx context.Context, userInput string) (call *function.Call, err error) {
	messages := []*MessageContent{
		{Role: "user", Content: userInput},
	}
	choices, err := r.completionWithToolClient.CompleteWithTool(ctx, messages, r.GetFuncDefs())
	if err != nil {
		return nil, err
	}
	if len(choices) < 1 {
		return nil, fmt.Errorf("no choices returned")
	}
	if len(choices[0].ToolCalls) < 1 {
		return nil, fmt.Errorf("no calls returned")
	}
	tc := choices[0].ToolCalls[0]
	fn, ok := r.fnName2fn[tc.Name]
	if !ok {
		return nil, fmt.Errorf("function %s not found", tc.Name)
	}
	info, err := fn.GetOrGenFuncInfo()
	if err != nil {
		return nil, err
	}
	rawParams := make(map[string]any)
	err = json.Unmarshal([]byte(tc.Args), &rawParams)
	if err != nil {
		return nil, err
	}
	call = &function.Call{
		Name:   tc.Name,
		Params: new(function.Params),
	}
	for _, p := range info.Params {

		//call.Params.Params = append(call.Params.Params, rawParams[p.Name])

		// strictly follow the reflection
		b, err := json.Marshal(rawParams[p.Name])
		if err != nil {
			return nil, err
		}
		call.Params.RawParams = append(call.Params.RawParams, string(b))
	}
	return
}

// resolveByPrompt resolves the user input to a function call just by prompt
func (r *Resolver) resolveByPrompt(ctx context.Context, userInput string) (call *function.Call, err error) {
	funcStr, err := r.getFuncStr(ctx, userInput)
	if err != nil {
		return nil, err
	}
	call, err = parseFuncStr(funcStr)
	return
}

func (r *Resolver) getFuncStr(ctx context.Context, userInput string) (string, error) {
	messages := []*MessageContent{
		{Content: r.sysPrompt, Role: "system"},
		{Content: userInput, Role: "user"},
	}
	choices, err := r.completionClient.Complete(ctx, messages)
	if err != nil {
		return "", err
	}
	if len(choices) < 1 {
		return "", fmt.Errorf("no choices returned")
	}
	funcStr := choices[0].Content
	return funcStr, nil
}

// parseFuncStr parses a function call string with name and given parameters like "funcName(param1, param2, ...)"
func parseFuncStr(funcStr string) (call *function.Call, err error) {
	matches := funcRex.FindStringSubmatch(funcStr)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid funcStr %s", funcStr)
	}
	// Get function name and parameters
	funcName := matches[1]
	rawParams := make([]string, 0)
	paramStr := matches[2]
	ps := make([]json.RawMessage, 0)
	s := fmt.Sprintf("[%s]", paramStr)
	if err = json.Unmarshal([]byte(s), &ps); err != nil {
		return nil, fmt.Errorf("invalid parameters to parse: %s", paramStr)
	}
	for _, p := range ps {
		rawParams = append(rawParams, string(p))
	}
	return &function.Call{
		Name: funcName,
		Params: &function.Params{
			RawParams: rawParams,
		},
	}, nil
}

func (r *Resolver) AddFunc(f *function.Function) bool {
	// add to store
	fName := f.GetName()
	if _, ok := r.fnName2fn[fName]; ok {
		return false
	}
	r.fnName2fn[fName] = f
	r.fnNames = append(r.fnNames, fName)

	// refresh sysPrompt
	if r.completionWithToolClient != nil {
		r.refreshSysPrompt()
	}
	return true
}

func (r *Resolver) GetFuncDefs() []*function.Definition {
	defs := make([]*function.Definition, 0, len(r.fnNames))
	for _, k := range r.fnNames {
		defs = append(defs, r.fnName2fn[k].GetDef())
	}
	return defs
}

func (r *Resolver) refreshSysPrompt() {
	funcDefs := make([]string, len(r.fnNames), len(r.fnNames))
	for idx, k := range r.fnNames {
		d, _ := json.Marshal(r.fnName2fn[k])
		funcDefs[idx] = string(d)
	}
	allDefStr := strings.Join(funcDefs, "\n")
	funcPrompt := fmt.Sprintf(r.sysPromptTemplate, allDefStr)
	r.sysPrompt = funcPrompt
}
