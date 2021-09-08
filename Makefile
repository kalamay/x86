INST_DIR:= asm/amd64/inst
INST_FILES:= $(wildcard $(INST_DIR)/*.go)
ENCODE_DIR:= test/encode
ENCODE_TESTS:= $(INST_FILES:$(INST_DIR)/%.go=$(ENCODE_DIR)/%_test.go)

tests: $(ENCODE_TESTS)
	go test ./...

$(ENCODE_DIR)/%_test.go: $(INST_DIR)/%.go ./test/gen.go
	go run ./test/gen.go -n $* -o $@
