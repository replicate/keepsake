# TODO: verify protoc is installed (brew install protobuf; go install google.golang.org/protobuf/cmd/protoc-gen-go)
# go get -u github.com/golang/protobuf/{proto,protoc-gen-go} google.golang.org/grpc
# make sure ~/go/bin is in $PATH, but make sure ~ is expanded! ("Strangely protoc can't expand ~" https://stackoverflow.com/a/57731186/135797)
# pip install grpcio
#
# TODO: make the process of installing the proto tools and verifying installation automatic.

GO_OUTPUT_DIR=../go/pkg/servicepb
PYTHON_OUTPUT_DIR=../python/replicate/servicepb
PROTO_NAME=replicate
SERVICE_PROTO=$(PROTO_NAME).proto

.PHONY: build
build:
	mkdir -p $(GO_OUTPUT_DIR)
	mkdir -p $(PYTHON_OUTPUT_DIR)
	protoc \
	  --go_out=$(GO_OUTPUT_DIR) \
	  --go_opt=paths=source_relative \
	  --go-grpc_out=$(GO_OUTPUT_DIR) \
	  --go-grpc_opt=paths=source_relative \
	  --experimental_allow_proto3_optional \
	  --mypy_out=$(PYTHON_OUTPUT_DIR) \
	  $(SERVICE_PROTO)
	python -m grpc_tools.protoc \
	  --python_out=$(PYTHON_OUTPUT_DIR) \
	  --grpc_python_out=$(PYTHON_OUTPUT_DIR) \
	  --proto_path=. \
	  --experimental_allow_proto3_optional \
	  $(SERVICE_PROTO)
	# need to post-process python-generated protobuf to make the import relative
	# TODO(andreas): remove this when https://github.com/protocolbuffers/protobuf/pull/7470 is merged
	sed -E -i '' 's/^import $(PROTO_NAME)_pb2 as $(PROTO_NAME)__pb2$$/from . import $(PROTO_NAME)_pb2 as $(PROTO_NAME)__pb2/' $(PYTHON_OUTPUT_DIR)/$(PROTO_NAME)_pb2_grpc.py
	touch $(PYTHON_OUTPUT_DIR)/__init__.py
