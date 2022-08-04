# Removes all the build directories
clean :
	-rm -r build
	-rm -r out

fmt :
	$(info Reformatting all source files...)
	go fmt ./...

build : clean fmt
	go build -o ./gen

test : build
	./gen \
		-out "./out" \
		-domain "example.com" \
		-gh-user "gentlecat"

run : build
	./gen \
		-out "./out"
