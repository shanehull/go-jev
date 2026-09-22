package jev

// Input is the state evaluated by the model. Construct it with State or Value.
type Input struct {
	value any
}

// State returns a text input.
func State(text string) Input {
	return Input{value: text}
}

// Value returns a structured input, which is sent as JSON.
func Value(v any) Input {
	return Input{value: v}
}
