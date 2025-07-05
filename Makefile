format:
	golangci-lint run --fix

lint:
	golangci-lint run

test:
	go test -v 

deploy:
	gcloud app deploy 

auth:
	gcloud auth login 