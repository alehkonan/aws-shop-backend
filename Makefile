.PHONY: seed deploy

STACK ?= --all

seed:
	go run seed/main.go

deploy:
	cdk deploy $(STACK)
