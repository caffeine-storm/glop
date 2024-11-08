package glew

// Define a subset of GLEW bindings

// #cgo darwin LDFLAGS: -framework OpenGL -lGLEW
// #cgo windows LDFLAGS: -lglew32 -lopengl32
// #cgo linux LDFLAGS: -lGLEW -lGL
// #cgo freebsd  CFLAGS: -I/usr/local/include
// #cgo freebsd LDFLAGS: -L/usr/local/lib -lglfw
// #include <GL/glew.h>
// static char const glop_glew_GL_ARB_imaging = GL_ARB_imaging;
import "C"

var (
	GL_ARB_imaging bool = C.GL_ARB_imaging == 0
)
