package jev

import (
	"encoding/json"
	"fmt"
)

// Usage reports token consumption for a request.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Response is the model's set of answers for one state.
type Response struct {
	Model   string            `json:"model"`
	Answers map[string]Answer `json:"answers"`
	Usage   Usage             `json:"usage"`
}

// Answer is a typed answer. Use a type switch, or the Choice, Score, and Noul
// accessors on Response.
type Answer interface {
	Kind() string
}

// ChoiceAnswer is the answer to a choice question.
type ChoiceAnswer struct {
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// Kind reports the answer type.
func (ChoiceAnswer) Kind() string { return "choice" }

// ScoreAnswer is the answer to a score question.
type ScoreAnswer struct {
	Score         float64            `json:"score"`
	Confidence    float64            `json:"confidence"`
	Legend        map[string]string  `json:"legend"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// Kind reports the answer type.
func (ScoreAnswer) Kind() string { return "score" }

// NoulAnswer is the answer to a noul question.
type NoulAnswer struct {
	Noul float64 `json:"noul"`
}

// Kind reports the answer type.
func (NoulAnswer) Kind() string { return "noul" }

// UnmarshalJSON decodes answers by their wire type discriminator.
func (r *Response) UnmarshalJSON(data []byte) error {
	var raw struct {
		Model   string                     `json:"model"`
		Answers map[string]json.RawMessage `json:"answers"`
		Usage   Usage                      `json:"usage"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("jev: decode response: %w", err)
	}

	r.Model = raw.Model
	r.Usage = raw.Usage
	r.Answers = make(map[string]Answer, len(raw.Answers))
	for name, payload := range raw.Answers {
		answer, err := decodeAnswer(payload)
		if err != nil {
			return fmt.Errorf("jev: answer %q: %w", name, err)
		}
		r.Answers[name] = answer
	}
	return nil
}

func decodeAnswer(payload json.RawMessage) (Answer, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(payload, &probe); err != nil {
		return nil, err
	}
	switch probe.Type {
	case "choice":
		var answer ChoiceAnswer
		if err := json.Unmarshal(payload, &answer); err != nil {
			return nil, err
		}
		return answer, nil
	case "score":
		var answer ScoreAnswer
		if err := json.Unmarshal(payload, &answer); err != nil {
			return nil, err
		}
		return answer, nil
	case "noul":
		var answer NoulAnswer
		if err := json.Unmarshal(payload, &answer); err != nil {
			return nil, err
		}
		return answer, nil
	default:
		return nil, fmt.Errorf("unknown type %q", probe.Type)
	}
}

// Choice returns the choice answer stored under name.
func (r *Response) Choice(name string) (ChoiceAnswer, bool) {
	answer, ok := r.Answers[name].(ChoiceAnswer)
	return answer, ok
}

// Score returns the score answer stored under name.
func (r *Response) Score(name string) (ScoreAnswer, bool) {
	answer, ok := r.Answers[name].(ScoreAnswer)
	return answer, ok
}

// Noul returns the noul answer stored under name.
func (r *Response) Noul(name string) (NoulAnswer, bool) {
	answer, ok := r.Answers[name].(NoulAnswer)
	return answer, ok
}
