# nlcall

👾 Call golang function by nature language, a demonstration.
```go
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
```

## How it works

* LLM to generate:
  * def doc
  * structured calling data
* reflection
* engineering

## Thanks
zomi team