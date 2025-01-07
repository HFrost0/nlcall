package function

import (
	"context"
	"fmt"
	"testing"
)

type MyStruct struct{}

func (s MyStruct) SayHello(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}

type Case struct {
	name      string
	f         any
	def       Definition
	ignoreIdx []int
	wantErr   bool
	CallCases []CallCase
}

type CallCase struct {
	name         string
	rawParams    []string
	ignoreParams []any
	result       []any
	wantErr      bool
}

func TestCreateFunction(t *testing.T) {
	tests := []Case{
		{
			name:      "test1",
			f:         func(ctx context.Context, a, b int) int { return a + b },
			ignoreIdx: []int{0},
			def: Definition{
				Name:        "add",
				Description: "add two numbers",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						// must be same order as function since go has no named parameters
						"a": map[string]any{"type": "number", "description": "first number"},
						"b": map[string]any{"type": "number", "description": "second number"},
					},
				},
			},
			CallCases: []CallCase{
				{
					name:         "call1",
					rawParams:    []string{"1", "2"},
					ignoreParams: []any{context.Background()},
					result:       []any{3},
					wantErr:      false,
				},
			},
		},
		{
			name:      "test2",
			f:         MyStruct{}.SayHello,
			ignoreIdx: []int{},
			def: Definition{
				Name:        "sayHello",
				Description: "say hello",
				//Parameters: &Parameters{
				//	Type: "object",
				//	Properties: map[string]Property{
				//		"name": {Type: "string", Description: "name"},
				//	},
				//},
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						// must be same order as function since go has no named parameters
						"name": map[string]any{"type": "string", "description": "name"},
					},
				},
			},
			CallCases: []CallCase{
				{
					name:         "call1",
					rawParams:    []string{`"world"`},
					ignoreParams: []any{},
					result:       []any{"Hello world!"},
					wantErr:      false,
				},
			},
		},
		{
			name: "test3",
			f: func(nums ...int) int {
				sum := 0
				for _, num := range nums {
					sum += num
				}
				return sum
			},
			ignoreIdx: []int{},
			def: Definition{
				Name:        "addMulti",
				Description: "add multiple numbers",
				//Parameters: &Parameters{
				//	Type: "object",
				//	Properties: map[string]Property{
				//		// must be same order as function since go has no named parameters
				//		"nums": {Type: "array", Description: "numbers"},
				//	},
				//},
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						// must be same order as function since go has no named parameters
						"nums": map[string]any{"type": "array", "description": "numbers"},
					},
				},
			},
			CallCases: []CallCase{
				{
					name:         "call1",
					rawParams:    []string{`[1,2,3]`},
					ignoreParams: []any{},
					result:       []any{6},
					wantErr:      false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := CreateFunction(tt.f, tt.def, tt.ignoreIdx...)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, callCase := range tt.CallCases {
				t.Run(callCase.name, func(t *testing.T) {
					res, err := f.Call(&Params{RawParams: callCase.rawParams}, callCase.ignoreParams...)
					if (err != nil) != callCase.wantErr {
						t.Errorf("Call() error = %v, wantErr %v", err, callCase.wantErr)
						return
					}
					if len(res) != len(callCase.result) {
						t.Errorf("Call() result = %v, want %v", res, callCase.result)
						return
					}
					for i := range res {
						if res[i] != callCase.result[i] {
							t.Errorf("Call() result = %v, want %v", res, callCase.result)
							return
						}
					}
				})
			}
		})
	}
}
