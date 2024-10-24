package render

import (
	"fmt"
	"log/slog"

	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/debug"
)

var shader_progs map[string]gl.Program

func init() {
	shader_progs = make(map[string]gl.Program)
}

type shaderError string

func (err shaderError) Error() string {
	return string(err)
}

// TODO(tmckee): refactor: There should be a 'DisableShader' to 'UseProgram(0)'
func EnableShader(name string) error {
	if name == "" {
		gl.Program(0).Use()
		return nil
	}
	prog, ok := shader_progs[name]
	if !ok {
		return shaderError(fmt.Sprintf("Tried to use unknown shader '%s'", name))
	}
	prog.Use()
	return nil
}

func SetUniformI(shader, variable string, n int) error {
	prog, ok := shader_progs[shader]
	if !ok {
		return shaderError(fmt.Sprintf("Tried to set a uniform in an unknown shader '%s'", shader))
	}
	loc := prog.GetUniformLocation(variable)
	loc.Uniform1i(n)
	return nil
}

func SetUniformF(shader, variable string, f float32) error {
	prog, ok := shader_progs[shader]
	if !ok {
		return shaderError(fmt.Sprintf("Tried to set a uniform in an unknown shader '%s'", shader))
	}
	loc := prog.GetUniformLocation(variable)
	loc.Uniform1f(f)
	return nil
}

func RegisterShader(name string, vertex, fragment string) error {
	if _, notOk := shader_progs[name]; notOk {
		return shaderError(fmt.Sprintf("Tried to register a shader called '%s' twice", name))
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

	shader_progs[name] = program

	debug.LogAndClearGlErrors(slog.Default())
	return nil
}
