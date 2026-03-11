package openai

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

type ResponseAPIInputParam struct {
	system, user *responses.ResponseInputItemUnionParam
	model        shared.ResponsesModel
}

func (p *ResponseAPIInputParam) ToRequestParam() responses.ResponseNewParams {
	return responses.ResponseNewParams{
		Model: p.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				*p.system,
				*p.user,
			},
		},
	}
}

type ResponseAPIInputParamOption func(*ResponseAPIInputParam)

func NewResponseAPIParam(opts ...ResponseAPIInputParamOption) *ResponseAPIInputParam {
	param := &ResponseAPIInputParam{}

	for _, opt := range opts {
		opt(param)
	}

	return param
}

func WithSystemMessage(content string) ResponseAPIInputParamOption {
	return func(param *ResponseAPIInputParam) {
		param.system = &responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: "system",
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: openai.String(content),
				},
			},
		}
	}
}

func WithUserMessage(content string) ResponseAPIInputParamOption {
	return func(param *ResponseAPIInputParam) {
		param.user = &responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: "user",
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: openai.String(content),
				},
			},
		}
	}
}

func WithModel(model string) ResponseAPIInputParamOption {
	return func(param *ResponseAPIInputParam) {
		param.model = shared.ChatModel(model)
	}
}
