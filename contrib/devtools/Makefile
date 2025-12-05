# DevTools Makefile for ShareHODL blockchain
# This file contains development and build tooling

###############################################################################
###                                Build Tools                             ###
###############################################################################

# Install development dependencies
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf files
proto-gen:
	@echo "Generating protobuf files..."
	buf generate

# Update protobuf dependencies  
proto-update-deps:
	@echo "Updating protobuf dependencies..."
	buf mod update

# Format protobuf files
proto-format:
	@echo "Formatting protobuf files..."
	find . -name "*.proto" -exec clang-format -i {} \;

.PHONY: install-tools proto-gen proto-update-deps proto-format