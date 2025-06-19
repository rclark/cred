cred:
	@go build -o cred main.go

clean:
	@rm -f cred

.PHONY: clean
