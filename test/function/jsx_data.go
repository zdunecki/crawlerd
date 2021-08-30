package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
	"rogchap.com/v8go"
)

// TODO: bundle react without node_modules
// TODO: speedup improvement (cache) - jsxData should not compile code every time even if jsx code didn't changed
func jsxData(jsxFile string, props map[string]string, out interface{}) error {
	var compiledCode []byte

	iifeVariable := "test"

	{
		result := api.Build(api.BuildOptions{
			EntryPoints: []string{jsxFile},
			Bundle:      true,
			Loader: map[string]api.Loader{
				".js":  api.LoaderJSX,
				".jsx": api.LoaderJSX,
			},
			GlobalName: iifeVariable,
			Format:     api.FormatIIFE,
		})

		if len(result.Errors) > 0 {
			return errors.New(result.Errors[0].Text)
		}

		compiledCode = result.OutputFiles[0].Contents
	}

	propsB, err := json.Marshal(props)
	if err != nil {
		return err
	}

	// v8
	ctx, err := v8go.NewContext()
	if err != nil {
		return err
	}

	if _, err := ctx.RunScript(string(compiledCode), "compiled_test.js"); err != nil {
		return err
	}

	//TODO: custom function support?
	//TODO: export default support?
	v, err := ctx.RunScript(fmt.Sprintf("%s.TestData(%s)", iifeVariable, propsB), "result.js")
	if err != nil {
		return err
	}

	vs, err := v.MarshalJSON()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(vs, &out); err != nil {
		return err
	}

	return nil
}
