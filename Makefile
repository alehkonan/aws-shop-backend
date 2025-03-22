include .env.local
export

STACK = $(word 2,$(MAKECMDGOALS))

build:
	go run cmd/build/main.go

seed:
	go run cmd/seed/main.go

test:
	go test -v ./...

deploy: build
	cdk deploy $(if $(STACK),$(STACK),--all) --debug

.PHONY: build seed test deploy