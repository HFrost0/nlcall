package main

import (
	"context"
	"fmt"
	"github.com/HFrost0/nlcall"
	"github.com/HFrost0/nlcall/llm"
	"log"
)

const dir = "./fn_def"

func main() {
	ctx := context.Background()
	client := NewLmClient("http://127.0.0.1:1234/v1/chat/completions", "qwen2.5-14b-instruct")

	agent := llm.NewLlmAgent(client)
	for _, f := range []any{
		add, greet, weather, lengthOfLongestSubstring, mul, no,
	} {
		if _, err := agent.RegisterFn(ctx, f, nlcall.WithLoadDefDir(dir)); err != nil {
			log.Fatal(err)
		}
	}
	fn, err := agent.AssignCallable(ctx, "1*3*34234*991238=?")
	if err != nil {
		log.Fatal(err)
	}
	res := fn()
	fmt.Println(res)
}
