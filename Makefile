gen-doc:
	@rm -f $$(find . -name README.mdx | paste -d ' ' -s -)
	@cd ./tools/compogen && go install .
	@go generate -run compogen ./...

gen-mock:
	@go install github.com/gojuno/minimock/v3/cmd/minimock@v3.3.13
	@go generate -run minimock ./...

test:
# Install tesseract via `brew install tesseract`
# Setup `export LIBRARY_PATH="/opt/homebrew/lib"` `export CPATH="/opt/homebrew/include"`
ifeq ($(shell uname), Darwin)
	@TESSDATA_PREFIX=$(shell dirname $(shell brew list tesseract | grep share/tessdata/eng.traineddata)) go test ./... -tags ocr
else
	@echo "This target can only be executed on Darwin (macOS)."
endif
