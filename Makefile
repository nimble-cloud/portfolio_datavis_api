prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/prod .

run:
	go build -o bin/app && ./bin/app