build-doc:
	@go install github.com/instill-ai/component/tools/compogen@latest

gen-doc:
	@rm -f $$(find . -name README.mdx | paste -d ' ' -s -)
	@go generate -run compogen ./...

# install tesseract via brew install tesseract
test:
ifeq ($(shell uname), Darwin)
	@TESSDATA_PREFIX=$(shell dirname $(shell brew list tesseract | grep share/tessdata/eng.traineddata)) go test ./... -tags ocr
else
	@echo "This target can only be executed on Darwin (macOS)."
endif
