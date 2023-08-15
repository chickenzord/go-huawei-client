package js

import (
	"encoding/json"
	"fmt"

	"github.com/robertkrimen/otto"
)

// Script
// Represents JavaScript code, need to be named for stacktrace purposes.
type Script struct {
	Name    string
	Content string
}

// EvalJSON evaluates the script content and call `fnCall` the return value will be JSON-stringified
// and unmarshaled back to `obj` with `json.Unmarshal`
func (s *Script) EvalJSON(result interface{}, fnCall string, args ...interface{}) error {
	vm := otto.New()

	if _, err := vm.Eval(s.Content); err != nil {
		return fmt.Errorf("error evaluating javascript: %w", err)
	}

	if _, err := vm.Eval(s.Content); err != nil {
		return fmt.Errorf("error evaluating script: %w", err)
	}

	v, err := vm.Call(fnCall, nil, args...)
	if err != nil {
		return fmt.Errorf("error calling %s: %w", fnCall, err)
	}

	bytes, err := v.MarshalJSON()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	return nil
}
