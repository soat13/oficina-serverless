build-IdentityFunction:
	@echo "ARTIFACTS_DIR=$(ARTIFACTS_DIR)"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags lambda.norpc -o $(ARTIFACTS_DIR)/bootstrap ./cmd/lambda
	@ls -la $(ARTIFACTS_DIR)