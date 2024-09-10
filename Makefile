build-doc:
	@go install github.com/instill-ai/component/tools/compogen@latest

gen-doc:
	@rm -f $$(find . -name README.mdx | paste -d ' ' -s -)
	@go generate -run compogen ./...

gen-mock:
	@go install github.com/gojuno/minimock/v3/cmd/minimock@v3.3.13
	@go generate -run minimock ./...

# For the future compogen developer, they can use this command to generate the documentation for the component by modified compogen tool.
# t stands for type of component and c stands for component name
# Example: make local-gen-doc t=application c=slack
local-gen-doc:
	@if [ -z "$(t)" ] && [ -z "$(c)" ]; then \
		cd ./tools/compogen && go install .; \
		cd ../..; \
		rm -f $$(find . -name README.mdx | paste -d ' ' -s -); \
		go generate -run compogen ./...; \
	elif [ -z "$(c)" ]; then \
		cd ./tools/compogen && go install .; \
		cd ../../${t}; \
		rm $$(find . -name README.mdx | paste -d ' ' -s -); \
		go generate -run compogen ./...; \
	else \
		cd ./tools/compogen && go install .; \
		cd ../..; \
		rm ${t}/${c}/v0/README.mdx; \
		go generate -run compogen ./${t}/${c}/v0; \
	fi

test:
# Install tesseract via `brew install tesseract`
# Setup `export LIBRARY_PATH="/opt/homebrew/lib"` `export CPATH="/opt/homebrew/include"`
ifeq ($(shell uname), Darwin)
	@TESSDATA_PREFIX=$(shell dirname $(shell brew list tesseract | grep share/tessdata/eng.traineddata)) go test ./... -tags ocr
else
	@echo "This target can only be executed on Darwin (macOS)."
endif
