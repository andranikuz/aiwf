package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	blogsdk "github.com/andranikuz/aiwf/examples/blog/sdk"
	"github.com/andranikuz/aiwf/runtime/go/aiwf"
)

type fakeClient struct{}

func (fakeClient) CallJSONSchema(_ context.Context, call aiwf.ModelCall) ([]byte, aiwf.Tokens, error) {
	switch call.OutputSchemaRef {
	case "aiwf://book/PremiseOutput":
		output := blogsdk.PremiseOutput{
			Logline: "Писатель находит портал во времени и спасает свой роман",
			Themes:  []string{"time-travel", "second-chances"},
			Tone:    blogsdk.ToneMysterious,
		}
		data, err := json.Marshal(output)
		return data, aiwf.Tokens{Prompt: 42, Completion: 128, Total: 170}, err
	case "aiwf://book/OutlineOutput":
		input, _ := call.Payload.(blogsdk.OutlineInput)
		output := blogsdk.OutlineOutput{
			Chapters: []blogsdk.OutlineOutputChaptersItem{
				{Id: "1", Title: "Завязка: появление портала"},
				{Id: "2", Title: "Герой исправляет ошибку прошлого"},
				{Id: "3", Title: fmt.Sprintf("Финал: публикация романа '%s'", input.Premise.Logline)},
			},
		}
		data, err := json.Marshal(output)
		return data, aiwf.Tokens{Prompt: 64, Completion: 256, Total: 320}, err
	default:
		return nil, aiwf.Tokens{}, fmt.Errorf("unsupported schema %q", call.OutputSchemaRef)
	}
}

func (fakeClient) CallJSONSchemaStream(context.Context, aiwf.ModelCall) (<-chan aiwf.StreamChunk, aiwf.Tokens, error) {
	ch := make(chan aiwf.StreamChunk)
	close(ch)
	return ch, aiwf.Tokens{}, errors.New("streaming not implemented")
}

func main() {
	ctx := context.Background()
	service := blogsdk.NewService(fakeClient{})

	result, trace, err := service.Workflows().Novel().Run(ctx, blogsdk.PremiseInput{})
	if err != nil {
		panic(err)
	}

	fmt.Println("Сгенерированный план глав:")
	for _, chapter := range result.Chapters {
		fmt.Printf("%s. %s\n", chapter.Id, chapter.Title)
	}
	if trace != nil {
		fmt.Printf("Всего токенов: %d\n", trace.Usage.Total)
	}
}
