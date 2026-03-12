package openai

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
	"github.com/sdkim96/indexing/part"
)

type ResponseAPIInputParam struct {
	system, user *responses.ResponseInputItemUnionParam
	model        shared.ResponsesModel
	format       *jsonschema.Schema
}

func (p *ResponseAPIInputParam) ToRequestParam() responses.ResponseNewParams {
	var respFormat responses.ResponseFormatTextConfigUnionParam
	if p.format != nil {
		b, _ := json.Marshal(p.format)
		var schemaMap map[string]any
		json.Unmarshal(b, &schemaMap)
		respFormat.OfJSONSchema = &responses.ResponseFormatTextJSONSchemaConfigParam{
			Name:   "response",
			Schema: schemaMap,
			Strict: openai.Bool(true),
		}
	} else {
		respFormat.OfText = &responses.ResponseFormatTextParam{}
	}

	return responses.ResponseNewParams{
		Model: p.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: []responses.ResponseInputItemUnionParam{
				*p.system,
				*p.user,
			},
		},
		Text: responses.ResponseTextConfigParam{
			Format: respFormat,
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
func WithPartsAsUserMessage(parts []part.Part) ResponseAPIInputParamOption {

	var openaiParts responses.ResponseInputMessageContentListParam

	for idx, p := range parts {
		openaiParts = append(openaiParts, responses.ResponseInputContentUnionParam{
			OfInputText: &responses.ResponseInputTextParam{
				Text: "Index: " + fmt.Sprintf("%d", idx) + " " + p.Text(),
			},
		})
	}

	return func(param *ResponseAPIInputParam) {
		var content string
		for _, p := range parts {
			content += p.Text() + "\n"
		}
		param.user = &responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: "user",
				Content: responses.EasyInputMessageContentUnionParam{
					OfInputItemContentList: openaiParts,
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

func WithResponseFormat[T any]() ResponseAPIInputParamOption {
	return func(p *ResponseAPIInputParam) {
		r := jsonschema.Reflector{
			DoNotReference: true,
		}
		schema := r.Reflect(new(T))
		p.format = schema.Definitions["T"]
		p.format = schema
	}
}
