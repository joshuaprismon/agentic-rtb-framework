#!/bin/bash
# Generate Go code from protobuf definitions
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$PROJECT_ROOT/proto"
OUT_DIR="$PROJECT_ROOT/pkg/pb"

# Ensure output directories exist
mkdir -p "$OUT_DIR/openrtb"
mkdir -p "$OUT_DIR/artf"

echo "Generating Go code from protobuf definitions..."

# Generate OpenRTB types
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  "$PROTO_DIR/com/iabtechlab/openrtb/v2.6/openrtb.proto"

# Generate ARTF service and types
protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/agenticrtbframework.proto"

echo "Done! Generated files in $OUT_DIR"
