package rendertest

import (
	"errors"
	"fmt"
	"image/color"
	"strings"

	"github.com/MobRulesGames/mathgl"
	"github.com/go-gl-legacy/gl"
	"github.com/runningwild/glop/glog"
	"github.com/runningwild/glop/render"
)

func getBadMatrixStackSizes() map[string]int {
	sizes := [3]int{
		gl.GetInteger(gl.MODELVIEW_STACK_DEPTH),
		gl.GetInteger(gl.PROJECTION_STACK_DEPTH),
		gl.GetInteger(gl.TEXTURE_STACK_DEPTH),
	}
	modes := [3]string{
		"modelview",
		"projection",
		"texture",
	}

	ret := map[string]int{}

	for i, sz := range sizes {
		if sz != 1 {
			ret[modes[i]] = sz
		}
	}

	return ret
}

func getBadMatrixValues() map[string]mathgl.Mat4 {
	var buffer [3]mathgl.Mat4

	gl.GetFloatv(gl.MODELVIEW_MATRIX, buffer[0][:])
	gl.GetFloatv(gl.PROJECTION_MATRIX, buffer[1][:])
	gl.GetFloatv(gl.TEXTURE_MATRIX, buffer[2][:])

	ret := map[string]mathgl.Mat4{}

	if !buffer[0].IsIdentity() {
		ret["modelview"] = buffer[0]
	}
	if !buffer[1].IsIdentity() {
		ret["projection"] = buffer[1]
	}
	if !buffer[2].IsIdentity() {
		ret["texture"] = buffer[2]
	}
	return ret

}

func checkMatrixInvariants() error {
	// If the matrix stacks are size 1 with the identity on top, something is
	// wrong.
	mp := getBadMatrixStackSizes()
	if len(mp) > 0 {
		return fmt.Errorf("matrix stacks needed to all be size 1: stack sizes: %+v", mp)
	}
	mpp := getBadMatrixValues()
	if len(mpp) > 0 {
		reports := []string{}
		for key, val := range mpp {
			reports = append(reports, fmt.Sprintf("%s:\n%v", key, render.Showmat(val)))
		}
		return fmt.Errorf("matrix stacks needed to be topped with identity matrices:\n%s", strings.Join(reports, "\n"))
	}

	return nil
}

func getImproperlyBoundState() []string {
	bindings := map[gl.GLenum]string{
		gl.ARRAY_BUFFER_BINDING:         "gl.ARRAY_BUFFER_BINDING",
		gl.ELEMENT_ARRAY_BUFFER_BINDING: "gl.ELEMENT_ARRAY_BUFFER_BINDING",
		gl.PIXEL_PACK_BUFFER_BINDING:    "gl.PIXEL_PACK_BUFFER_BINDING",
		gl.PIXEL_UNPACK_BUFFER_BINDING:  "gl.PIXEL_UNPACK_BUFFER_BINDING",
		gl.TEXTURE_BINDING_2D:           "gl.TEXTURE_BINDING_2D",
	}

	badvals := []string{}
	for code, name := range bindings {
		val := gl.GetInteger(code)
		if val != 0 {
			badvals = append(badvals, name)
		}
	}

	return badvals
}

func checkBindingsInvariants() error {
	badvals := getImproperlyBoundState()
	if len(badvals) == 0 {
		return nil
	}

	return fmt.Errorf("need bindings unset but found bindings for: %v", badvals)
}

func checkColourInvariants() error {
	// default fg/bg is white on black
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	fg := GetCurrentForegroundColour()
	errs := []error{}
	if fg != white {
		errs = append(errs, fmt.Errorf("bad foreground colour: %v expected: %v (white)", fg, white))
	}

	bg := GetCurrentBackgroundColor()
	// Alpha doesn't matter for background/clearing
	bg.A = black.A
	if bg != black {
		errs = append(errs, fmt.Errorf("bad background colour: %v expected %v (black)", bg, black))
	}

	return errors.Join(errs...)
}

func checkInvariants() error {
	return errors.Join(
		checkMatrixInvariants(),
		checkBindingsInvariants(),
		checkColourInvariants(),
	)
}

func enforceMatrixStacksMustBeIdentitySingletons() {
	sizes := [3]int{
		gl.GetInteger(gl.MODELVIEW_STACK_DEPTH),
		gl.GetInteger(gl.PROJECTION_STACK_DEPTH),
		gl.GetInteger(gl.TEXTURE_STACK_DEPTH),
	}
	modes := [3]render.MatrixMode{
		render.MatrixModeModelView,
		render.MatrixModeProjection,
		render.MatrixModeTexture,
	}

	for i, sizei := range sizes {
		if sizei != 1 {
			glog.WarningLogger().Warn("rendertest enforcing matrix invariant", "state leakage", fmt.Sprintf("matrix mode %v", modes[i]))
		}
		gl.MatrixMode(gl.GLenum(modes[i]))
		for j := sizei; j > 1; j-- {
			gl.PopMatrix()
		}
	}

	badMats := getBadMatrixValues()
	if len(badMats) > 0 {
		glog.WarningLogger().Warn("rendertest enforcing matrix invariant", "state leakage", "one or more matrices had non-identity value", "variants", badMats)
	}

	for i := range sizes {
		gl.MatrixMode(gl.GLenum(modes[i]))
		gl.LoadIdentity()
	}
}

func enforceClearBindingsSet() {
	badBindings := getImproperlyBoundState()
	if len(badBindings) > 0 {
		glog.WarningLogger().Warn("rendertest enforcing bindings invariant", "state leakage", badBindings)
	}

	bufferBindings := []gl.GLenum{
		gl.ARRAY_BUFFER,
		gl.ELEMENT_ARRAY_BUFFER,
		gl.PIXEL_PACK_BUFFER,
		gl.PIXEL_UNPACK_BUFFER,
	}
	for _, name := range bufferBindings {
		gl.Buffer(0).Bind(name)
	}

	textureBindings := []gl.GLenum{
		gl.TEXTURE_2D,
	}
	for _, name := range textureBindings {
		gl.Texture(0).Bind(name)
	}
}

func enforceColourInvariants() {
	gl.Color4f(1, 1, 1, 1)
	gl.ClearColor(0, 0, 0, 1)
}

func enforceInvariants() {
	enforceMatrixStacksMustBeIdentitySingletons()
	enforceClearBindingsSet()
	enforceColourInvariants()
}
