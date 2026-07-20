OUTPUTS_FILE := cdk-outputs.json

.PHONY: bootstrap build-api deploy diff destroy

bootstrap:
	cdk bootstrap

build-api:
	mkdir -p ../api/dist
	cd ../api && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/bootstrap .

deploy: build-api
	cdk deploy --outputs-file $(OUTPUTS_FILE) --require-approval never --import-existing-resources
	go run ./gen-config

diff:
	cdk diff

destroy:
	cdk destroy
