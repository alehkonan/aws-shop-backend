.PHONY: build seed test deploy

STACK = $(word 2,$(MAKECMDGOALS))

build:
	go run cmd/build/main.go

seed:
	go run cmd/seed/main.go

test:
	go test -v ./...

deploy: build
	cdk deploy $(if $(STACK),$(STACK),--all)
