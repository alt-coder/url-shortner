PROTO_DIR := proto

.PHONY: proto
proto:
	protoc -I $(PROTO_DIR) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_DIR)/url_shortener.proto

.PHONY: clean
clean:
	rm -f $(PROTO_DIR)/*.pb.go
	rm -f $(PROTO_DIR)/*.gw.go
