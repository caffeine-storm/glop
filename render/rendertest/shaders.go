package rendertest

import (
	"errors"
	"fmt"

	"github.com/runningwild/glop/render"
)

func WithShaderProgs(shaders *render.ShaderBank, vertShader string, fragShader string, fn func()) {
	err := shaders.RegisterShader("debugshaders", vertShader, fragShader)
	if err != nil && !errors.Is(err, render.ErrShaderAlreadyRegistered) {
		// If 'debugshaders' is already registered, that's fine.
		panic(fmt.Errorf("couldn't register debug shaders: %w", err))
	}

	err = shaders.EnableShader("debugshaders")
	if err != nil {
		panic(fmt.Errorf("couldn't enable debug shaders: %w", err))
	}

	defer func() {
		shaders.EnableShader("")
	}()
	fn()
}
