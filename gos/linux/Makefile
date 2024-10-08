all: lib/libglop.so

lib/libglop.so: glop.cpp include/glop.h
	mkdir -p lib/
	${CXX} -Wall -fPIC -Iinclude glop.cpp -shared -o $@

clean:
	${RM} -f lib/libglop.so

.PHONY: clean
