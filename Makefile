test:
	go test -v

cover:
	rm -rf *.coverprofile
	go test -coverprofile=compress.coverprofile
	gover
	go tool cover -html=compress.coverprofile