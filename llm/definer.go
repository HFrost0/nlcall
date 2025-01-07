package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HFrost0/nlcall/function"
)

var fnDefSysPromptTemplate = `Your task is to output a formatted json string that defines a golang function. you will receive a json string like:
{
	"name": "<fn_name>",
	"comments": "<fn_comments>",
	"source_code": "<fn_source_code>"
}
try to understand the info, your output should be a informative json string like:
'''
{"name":"greet","description":"return a person's greeting with his/her name and age. Call string example: greet(\"李宁\",15) or greet(\"jack\",14)","parameters":{"properties":{"age":{"description":"the person's age","type":"integer"},"name":{"description":"the person's name","type":"string"}},"type":"object"}}
{"name":"no","description":"use this func if there's no suitable function for user's input or user's request is unrelated to the existed funcs. Calling example: no()","parameters":null}
{"name":"add","description":"return the sum of integers. Calling example: add([1,1]) add([1,2,4]). be aware that the input must be a list of integers","parameters":{"properties":{"nums":{"description":"multiple integers which will be added together","type":"integers"}},"type":"object"}}
'''
follow the rules:
1. output without any explanation.
2. output the json string in one line.
`

type Definer struct {
	completionClient CompletionClient
	systemPrompt     string
}

func NewDefiner(completionClient CompletionClient) *Definer {
	return &Definer{
		systemPrompt:     fnDefSysPromptTemplate,
		completionClient: completionClient,
	}
}

func (l *Definer) Define(ctx context.Context, fn any) (*function.Definition, error) {
	fnInfo, err := function.GetFunctionDetails(fn)
	if err != nil {
		return nil, err
	}
	funcMsg, err := json.Marshal(fnInfo)
	if err != nil {
		return nil, err
	}
	choices, err := l.completionClient.Complete(ctx, []*MessageContent{
		{Role: "system", Content: l.systemPrompt},
		{Role: "user", Content: string(funcMsg)},
	})
	if err != nil {
		return nil, err
	}
	if len(choices) < 1 {
		return nil, fmt.Errorf("no choices returned")
	}
	respStr := choices[0].Content
	// 创建函数定义
	res := new(function.Definition)
	err = json.Unmarshal([]byte(respStr), res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
