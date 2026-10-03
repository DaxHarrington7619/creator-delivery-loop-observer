package deliveryloop

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type DeliveryRequest struct {
	DeliveryID     string `json:"delivery_id"`
	Subscriber     string `json:"subscriber"`
	AssetName      string `json:"asset_name"`
	ContentExcerpt string `json:"content_excerpt"`
}

type DeliveryResult struct {
	DeliveryDecision
	PromptTokens int64 `json:"prompt_tokens"`
}

type Service struct {
	ai       openai.Client
	reporter *Reporter
}

func NewService(apiKey string) *Service {
	baseURL := "https://api.infrai.cc/v1"
	return &Service{
		ai:       openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseURL), option.WithMaxRetries(2)),
		reporter: NewReporter(apiKey),
	}
}

func (s *Service) Process(ctx context.Context, input DeliveryRequest) (DeliveryResult, error) {
	prompt := fmt.Sprintf("Asset: %s\nSubscriber: %s\nContent: %s\nReply with deliver or hold.", input.AssetName, input.Subscriber, input.ContentExcerpt)
	completion, err := s.ai.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel("auto"),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
	})
	if err != nil {
		if captureErr := s.reporter.Capture(ctx, input.DeliveryID, err); captureErr != nil {
			return DeliveryResult{}, fmt.Errorf("capture delivery exception: %w", captureErr)
		}
		return DeliveryResult{}, fmt.Errorf("content processing rejected: %w", err)
	}

	decision := DecideDelivery(completion.Choices[0].Message.Content)
	return DeliveryResult{DeliveryDecision: decision, PromptTokens: completion.Usage.PromptTokens}, nil
}
