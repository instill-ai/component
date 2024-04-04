build-doc:
	go install github.com/instill-ai/component/tools/compogen@latest
gen-doc:
	@rm -f $$(find . -name README.mdx | paste -d ' ' -s -)
	@go generate -run compogen ./...
