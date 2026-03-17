// Copyright 2026 Sungdong Kim
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openai

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
	"github.com/sdkim96/indexing/cache"
	"github.com/sdkim96/indexing/part"
)

// ResponseAPIParam is a struct that holds parameters for the OpenAI API request,
// when creating a new response.
// It includes:
//   - system and user messages
//   - the model to be used
//   - the expected response format.
type ResponseAPIParam struct {
	system, user    *responses.ResponseInputItemUnionParam
	model           shared.ResponsesModel
	format          *jsonschema.Schema
	reasoningEffort shared.ReasoningEffort
	temperature     float64
}

func (p *ResponseAPIParam) System() *responses.ResponseInputItemUnionParam {
	return p.system
}

func (p *ResponseAPIParam) User() *responses.ResponseInputItemUnionParam {
	return p.user
}

func (p *ResponseAPIParam) Model() shared.ResponsesModel {
	return p.model
}

func (p *ResponseAPIParam) Format() *jsonschema.Schema {
	return p.format
}

func (p *ResponseAPIParam) ReasoningEffort() shared.ReasoningEffort {
	return p.reasoningEffort
}

func (p *ResponseAPIParam) Temperature() float64 {
	return p.temperature
}

func (p *ResponseAPIParam) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		System      *responses.ResponseInputItemUnionParam `json:"system,omitempty"`
		User        *responses.ResponseInputItemUnionParam `json:"user,omitempty"`
		Model       shared.ResponsesModel                  `json:"model,omitempty"`
		Format      *jsonschema.Schema                     `json:"format,omitempty"`
		Reasoning   shared.ReasoningEffort                 `json:"reasoning_effort,omitempty"`
		Temperature float64                                `json:"temperature,omitempty"`
	}{
		System:      p.system,
		User:        p.user,
		Model:       p.model,
		Format:      p.format,
		Reasoning:   p.reasoningEffort,
		Temperature: p.temperature,
	})
}

var _ cache.Hasher = (*ResponseAPIParam)(nil)

func (p *ResponseAPIParam) FingerPrint(prefix string) string {
	b, _ := json.Marshal(p)
	key := cache.Sha256(b)
	if prefix != "" {
		key = prefix + ":" + key
	}
	return key
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

	params := responses.ResponseNewParams{
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
		Reasoning: responses.ReasoningParam{
			Effort: p.reasoningEffort,
		},
	}
	if p.temperature != 0 {
		params.Temperature = openai.Float(p.temperature)
	}

	return params
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

func WithReasoningEffort(level string) ResponseAPIParamOption {
	return func(p *ResponseAPIParam) {
		switch level {
		case "low":
			p.reasoningEffort = responses.ReasoningEffortLow
		case "medium":
			p.reasoningEffort = responses.ReasoningEffortMedium
		case "high":
			p.reasoningEffort = responses.ReasoningEffortHigh
		default:
			p.reasoningEffort = responses.ReasoningEffortMedium
		}
	}
}

func WithTemperature(temp float64) ResponseAPIParamOption {
	return func(p *ResponseAPIParam) {
		p.temperature = temp
	}
}

type EmbeddingAPIParam struct {
	input string
	model string
}

func (p *EmbeddingAPIParam) Input() string {
	return p.input
}

func (p *EmbeddingAPIParam) Model() string {
	return p.model
}

func (p *EmbeddingAPIParam) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Input string `json:"input,omitempty"`
		Model string `json:"model,omitempty"`
	}{
		Input: p.input,
		Model: p.model,
	})
}

func (p *EmbeddingAPIParam) FingerPrint(prefix string) string {
	b, _ := json.Marshal(p)
	key := cache.Sha256(b)
	if prefix != "" {
		key = prefix + ":" + key
	}
	return key
}

var _ cache.Hasher = (*EmbeddingAPIParam)(nil)

func NewEmbeddingAPIParam(opts ...EmbeddingAPIParamOption) *EmbeddingAPIParam {
	p := &EmbeddingAPIParam{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type EmbeddingAPIParamOption func(*EmbeddingAPIParam)

func WithEmbeddingInput(input string) EmbeddingAPIParamOption {
	return func(p *EmbeddingAPIParam) {
		p.input = input
	}
}

func WithEmbeddingModel(model string) EmbeddingAPIParamOption {
	return func(p *EmbeddingAPIParam) {
		p.model = model
	}
}
