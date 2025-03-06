.PHONY: build seed deploy

STACK = $(word 2,$(MAKECMDGOALS))

build:
	go run cmd/build/main.go

seed:
	go run cmd/seed/main.go

deploy: build
	cdk deploy $(if $(STACK),$(STACK),--all)
