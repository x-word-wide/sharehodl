#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code..."

# Get the proto directory
PROTO_DIR="./proto"

# Output directories
API_DIR="./api"

# Create output directories
mkdir -p "$API_DIR"

# Generate code for each module
for module in equity dex governance hodl validator explorer; do
    MODULE_PROTO="$PROTO_DIR/sharehodl/$module"
    if [ -d "$MODULE_PROTO" ]; then
        echo "Generating protos for $module module..."

        # Generate Go code with gRPC gateway
        buf generate --template buf.gen.gogo.yaml --path "$MODULE_PROTO" 2>/dev/null || \
        protoc \
            -I "$PROTO_DIR" \
            -I "$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)/proto" \
            -I "$(go list -f '{{ .Dir }}' -m github.com/cosmos/gogoproto)" \
            -I "$(go list -f '{{ .Dir }}' -m github.com/googleapis/googleapis)" \
            --gogo_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
            --grpc-gateway_out=logtostderr=true,allow_colon_final_segments=true:. \
            $(find "$MODULE_PROTO" -name "*.proto") 2>/dev/null || echo "Warning: Could not generate protos for $module"
    fi
done

echo "Proto generation complete!"
