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

// ResponseAPIParam is a struct that holds parameters for the OpenAI API request,
// when creating a new response.
// It includes:
//   - system and user messages
//   - the model to be used
//   - the expected response format.
type ResponseAPIParam struct {
	system, user *responses.ResponseInputItemUnionParam
	model        shared.ResponsesModel
	format       *jsonschema.Schema
}

// NewResponseAPIParam creates a new ResponseAPIParam with the given options,
// providing an easy way to set various parameters for the OpenAI API request.
func NewResponseAPIParam(opts ...ResponseAPIParamOption) *ResponseAPIParam {
	param := &ResponseAPIParam{}

	for _, opt := range opts {
		opt(param)
	}

	return param
}

// ToRequestParam converts the ResponseAPIParam to a format suitable for making an API request to OpenAI,
// specifically to the responses endpoint. It constructs the request parameters based on the system and user messages,
// the model, and the expected response format.
func (p *ResponseAPIParam) ToRequestParam() responses.ResponseNewParams {
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

type ResponseAPIParamOption func(*ResponseAPIParam)

func WithSystemMessage(content string) ResponseAPIParamOption {
	return func(param *ResponseAPIParam) {
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

func WithUserMessage(content string) ResponseAPIParamOption {
	return func(param *ResponseAPIParam) {
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
func WithPartsAsUserMessage(parts []part.Part) ResponseAPIParamOption {

	var openaiParts responses.ResponseInputMessageContentListParam

	for idx, p := range parts {
		openaiParts = append(openaiParts, responses.ResponseInputContentUnionParam{
			OfInputText: &responses.ResponseInputTextParam{
				Text: "Index: " + fmt.Sprintf("%d", idx) + " " + p.Text(),
			},
		})
	}

	return func(param *ResponseAPIParam) {
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

func WithModel(model string) ResponseAPIParamOption {
	return func(param *ResponseAPIParam) {
		param.model = shared.ChatModel(model)
	}
}

func WithResponseFormat[T any]() ResponseAPIParamOption {
	return func(p *ResponseAPIParam) {
		r := jsonschema.Reflector{
			DoNotReference: true,
		}
		schema := r.Reflect(new(T))
		p.format = schema.Definitions["T"]
		p.format = schema
	}
}
