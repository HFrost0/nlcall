package llm

import "context"

//type ResolverClient interface {
//	CompletionClient | FunctionCallClient
//}

type CompletionClient interface {
	Complete(ctx context.Context, messages []*MessageContent) ([]*ChoiceContent, error)
}

type CompletionWithToolClient interface {
	CompletionClient
	CompleteWithTool(ctx context.Context, messages []*MessageContent, tools []*Tool) ([]*ChoiceContent, error)
}
