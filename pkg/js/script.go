package js

import (
	"encoding/json"
	"fmt"

	"rogchap.com/v8go"
)

type Script struct {
	Name    string
	Content string
}

func (s *Script) EvalJSON(fnCall string, obj interface{}) error {
	ctx := v8go.NewContext()
	if _, err := ctx.RunScript(s.Content, s.Name); err != nil {
		return err
	}

	val, err := ctx.RunScript(fmt.Sprintf("JSON.stringify(%s)", fnCall), "json_stringify.js")
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val.String()), &obj); err != nil {
		return err
	}

	return nil
}
