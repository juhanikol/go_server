.PHONY: fmt test vet vuln check

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

vuln:
	govulncheck ./...

check: fmt test vet vuln