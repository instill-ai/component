build-doc:
	@go install github.com/instill-ai/component/tools/compogen@latest

gen-doc:
	@rm -f $$(find . -name README.mdx | paste -d ' ' -s -)
	@go generate -run compogen ./...

gen-mock:
	@go install github.com/gojuno/minimock/v3/cmd/minimock@v3.3.13
	@go generate -run minimock ./...

# t stands for type of component and c stands for component name
# Example: make gen-doc-test t=application c=slack
gen-doc-test:
	@cd ./tools/compogen && go install .
	@rm ./${t}/${c}/v0/README.mdx
	@go generate -run compogen ./${t}/${c}/v0

test:
# Install tesseract via `brew install tesseract`
# Setup `export LIBRARY_PATH="/opt/homebrew/lib"` `export CPATH="/opt/homebrew/include"`
ifeq ($(shell uname), Darwin)
	@TESSDATA_PREFIX=$(shell dirname $(shell brew list tesseract | grep share/tessdata/eng.traineddata)) go test ./... -tags ocr
else
	@echo "This target can only be executed on Darwin (macOS)."
endif
