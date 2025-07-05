format:
	golangci-lint run --fix

test:
	go test -v 

deploy:
	gcloud app deploy 

auth:
	gcloud auth login 