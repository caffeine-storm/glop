
all: drawstring

drawstring: main.go go.mod go.sum
	go build -x -o $@ -ldflags "-extldflags \"-L ../../gos/linux/lib -Wl,-rpath,\$${ORIGIN}/../../gos/linux/lib\"" .
