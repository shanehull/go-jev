package jev

import "encoding/json"

// Questions maps a question name to its definition.
type Questions map[string]Question

// Question is a typed question for the model. Construct it with Choice, Score,
// or Noul.
type Question struct {
	Type         string          `json:"type"`
	Instructions string          `json:"instructions"`
	Criteria     json.RawMessage `json:"criteria,omitempty"`
}

// Choice builds a choice question. The model picks one option and returns a
// probability for each option.
func Choice(instructions string, criteria map[string]string) Question {
	return Question{
		Type:         "choice",
		Instructions: instructions,
		Criteria:     mustMarshal(criteria),
	}
}

// Score builds a score question. The model places the state on an ordered rubric.
func Score(instructions string, criteria []string) Question {
	return Question{
		Type:         "score",
		Instructions: instructions,
		Criteria:     mustMarshal(criteria),
	}
}

// Noul builds a noul question. The model returns the probability that the
// statement is true.
func Noul(instructions string) Question {
	return Question{Type: "noul", Instructions: instructions}
}

func mustMarshal(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return raw
}
