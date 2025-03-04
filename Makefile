.PHONY: seed deploy

STACK ?= --all

seed:
	go run cmd/seed/main.go

deploy:
	cdk deploy $(STACK)
