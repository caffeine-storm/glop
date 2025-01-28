package render

import (
	"fmt"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
	"github.com/runningwild/glop/glog"
)

type ShaderBank struct {
	ShaderProgs map[string]gl.Program
}

func MakeShaderBank() *ShaderBank {
	return &ShaderBank{
		ShaderProgs: make(map[string]gl.Program),
	}
}

func (bank *ShaderBank) HasShader(shaderName string) bool {
	_, found := bank.ShaderProgs[shaderName]
	return found
}

type shaderError string

func (err shaderError) Error() string {
	return string(err)
}

var ErrShaderAlreadyRegistered shaderError = "shader name already in use"

// TODO(tmckee): refactor: There should be a 'DisableShader' to 'UseProgram(0)'
func (bank *ShaderBank) EnableShader(name string) error {
	if name == "" {
		gl.Program(0).Use()
		return nil
	}
	prog, ok := bank.ShaderProgs[name]
	if !ok {
		return shaderError(fmt.Sprintf("Tried to use unknown shader '%s'", name))
	}
	prog.Use()
	debug.LogAndClearGlErrors(glog.DebugLogger())
	return nil
}

func (bank *ShaderBank) SetUniformI(shader, variable string, n int) error {
	prog, ok := bank.ShaderProgs[shader]
	if !ok {
		return shaderError(fmt.Sprintf("Tried to set a uniform in an unknown shader '%s'", shader))
	}
	loc := prog.GetUniformLocation(variable)
	loc.Uniform1i(n)
	return nil
}

func (bank *ShaderBank) SetUniformF(shader, variable string, f float32) error {
	prog, ok := bank.ShaderProgs[shader]
	if !ok {
		return shaderError(fmt.Sprintf("Tried to set a uniform in an unknown shader '%s'", shader))
	}
	loc := prog.GetUniformLocation(variable)
	loc.Uniform1f(f)
	return nil
}

func (bank *ShaderBank) RegisterShader(name string, vertex, fragment string) error {
	if _, notOk := bank.ShaderProgs[name]; notOk {
		return ErrShaderAlreadyRegistered
	}

	vertex_shader := gl.CreateShader(gl.VERTEX_SHADER)
	vertex_shader.Source(vertex)
	vertex_shader.Compile()
	did_compile := vertex_shader.Get(gl.COMPILE_STATUS)
	if did_compile != gl.TRUE {
		return shaderError(fmt.Sprintf("Failed to compile vertex shader '%s': %v", name, did_compile))
	}

	fragment_shader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fragment_shader.Source(fragment)
	fragment_shader.Compile()
	did_compile = fragment_shader.Get(gl.COMPILE_STATUS)
	if did_compile != gl.TRUE {
		return shaderError(fmt.Sprintf("Failed to compile fragment shader '%s': %v", name, did_compile))
	}

	// shader successfully compiled - now link
	program := gl.CreateProgram()
	program.AttachShader(vertex_shader)
	program.AttachShader(fragment_shader)
	program.Link()
	did_link := program.Get(gl.LINK_STATUS)
	if did_link != gl.TRUE {
		return shaderError(fmt.Sprintf("Failed to link shader '%s': %v", name, did_compile))
	}

	bank.ShaderProgs[name] = program

	debug.LogAndClearGlErrors(glog.InfoLogger())
	return nil
}
