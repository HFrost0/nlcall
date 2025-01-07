package llm

import "github.com/HFrost0/nlcall"

func NewLlmAgent(client CompletionClient) *nlcall.Agent {
	resolver := NewResolver(client)
	definer := NewDefiner(client)
	return nlcall.NewAgent(resolver, definer)
}
