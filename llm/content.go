package llm

import "github.com/HFrost0/nlcall/function"

type MessageContent struct {
	Role    string
	Content string
}

type ChoiceContent struct {
	Content   string
	ToolCalls []*ToolCall
}

type ToolCall struct {
	Name string
	Args string // should be a json string
}

type Tool = function.Definition
