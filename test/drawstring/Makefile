all: it

it: drawstring

go: drawstring
	./$^

drawstring: main.go go.mod go.sum
	go build -x -o $@ -ldflags "-extldflags \"-L ../../gos/linux/lib -Wl,-rpath,\$${ORIGIN}/../../gos/linux/lib\"" .

.PHONY: it
.PHONY: go
.PHONY: drawstring
