CXX=i686-w64-mingw32-g++-win32
LD=i686-w64-mingw32-ld
CXXFLAGS=-m32
CPPFLAGS=-Iinclude
LDFLAGS=-shared -lopengl32 -lgdi32 -ldxguid -lwinmm -ldinput

all: lib/libglop.dll

glop.o: |include/glop.h
glop.o: glop.cpp
	${CXX} ${CPPFLAGS} ${CXXFLAGS} -o $@ -c $^

lib/libglop.dll: glop.o |lib
	${CXX} ${CPPFLAGS} -o $@ $^ ${LDFLAGS}

lib:
	mkdir -p $@

clean:
	rm -f lib/libglop.dll glop.o

.PHONY: clean
