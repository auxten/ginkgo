.PHONY: all compile
all: compile

compile:
	go build cmd/ginkgo.go
