.PHONY: run
run:
	go run ./main.go

.PHONY: build-win
build-win:
	GOOS=windows GOARCH=amd64 go build ./main.go
	