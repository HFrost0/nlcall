package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/HFrost0/nlcall/function"
	"github.com/HFrost0/nlcall/llm"
	"io"
	"net/http"
)

type LmClient struct {
	*http.Client
	Model       string
	Headers     map[string]string
	Url         string
	ReqMaxRetry int
}

func NewLmClient(url string, model string) *LmClient {
	return &LmClient{
		Url:         url,
		ReqMaxRetry: 1,
		Client:      &http.Client{},
		Model:       model,
		Headers: map[string]string{
			"Content-Type":    "application/json",
			"GetContent-Type": "application/json",
		},
	}
}

func (r *LmClient) Complete(ctx context.Context, messages []*llm.MessageContent) ([]*llm.ChoiceContent, error) {
	body, err := r.ToBytes(messages, nil)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 1+r.ReqMaxRetry; i++ {
		respBytes, err := r.sendRequest(ctx, body)
		if err != nil {
			continue
		}
		msgContent, err := r.FromBytes(respBytes)
		if err != nil {
			continue
		}
		return msgContent, nil
	}
	return nil, fmt.Errorf("requester failed after %d attempts", 1+r.ReqMaxRetry)
}

func (r *LmClient) CompleteWithTool(ctx context.Context, messages []*llm.MessageContent, tools []*llm.Tool) ([]*llm.ChoiceContent, error) {
	body, err := r.ToBytes(messages, tools)
	if err != nil {
		return nil, err
	}
	for i := 0; i < 1+r.ReqMaxRetry; i++ {
		respBytes, err := r.sendRequest(ctx, body)
		if err != nil {
			continue
		}
		msgContent, err := r.FromBytes(respBytes)
		if err != nil {
			continue
		}
		return msgContent, nil
	}
	return nil, fmt.Errorf("requester %s failed after %d attempts", 1+r.ReqMaxRetry)
}

func (r *LmClient) sendRequest(ctx context.Context, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", r.Url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}
	resp, err := r.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requester response status code %d", resp.StatusCode)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

type LmReq struct {
	Model       string     `json:"model"`
	Tools       []*Tool    `json:"tools"`
	Messages    []*Message `json:"messages"`
	Temperature float64    `json:"temperature"`
	Stream      bool       `json:"stream"`
}

type Tool struct {
	Type     string               `json:"type"`
	Function *function.Definition `json:"function"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type LmResp struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func (r *LmClient) ToBytes(messages []*llm.MessageContent, tools []*llm.Tool) ([]byte, error) {
	req := &LmReq{
		Model:       r.Model,
		Temperature: 1.0,
	}
	for _, msg := range messages {
		req.Messages = append(req.Messages, &Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	for _, tool := range tools {
		req.Tools = append(req.Tools, &Tool{
			Type:     "function",
			Function: tool,
		})
	}
	return json.Marshal(req)
}

func (r *LmClient) FromBytes(bytes []byte) ([]*llm.ChoiceContent, error) {
	resp := &LmResp{}
	if err := json.Unmarshal(bytes, resp); err != nil {
		return nil, err
	}
	var choices []*llm.ChoiceContent
	for _, c := range resp.Choices {
		choice := &llm.ChoiceContent{
			Content: c.Message.Content,
		}
		for _, tc := range c.Message.ToolCalls {
			toolCall := &llm.ToolCall{
				Name: tc.Function.Name,
				Args: tc.Function.Arguments,
			}
			choice.ToolCalls = append(choice.ToolCalls, toolCall)
		}
		choices = append(choices, choice)
	}
	return choices, nil
}
