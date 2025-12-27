Building a docker image:  
`docker build -t kuard .`

Running a docker image:  
`docker run --rm -p 8080:8080 kuard`

Limiting memory and CPU:  
```
docker run -d --name kuard \
  --publish 8080:8080 \
  --memory 200m \
  --memory-swap 1G \
  --cpu-shares 1024 \
  gcr.io/kuar-demo/kuard-amd64:blue
```

Publishing a docker image:  
Can be published to Google Container Registry (GCR), Amazon Elastic Container Registry (ECR), or Docker Hub  

```
docker tag guard gcr.io/kuar-demo/kuard-amd64:blue
docker push gcr.io/kuar-demo/kuard-amd64:blue
```

Deploying a published container:  
```
docker run -d --name kuard \
  --publish 8080:8080 \
  gcr.io/kuar-demo/kuard-amd64:blue
```

Deleting a docker image:  
`docker rmi <image-id-or-tag-name>`

List docker images:  
`docker images`

General cleanup:  
Removes untagged images, stopped containers, unused cache.
`docker system prune`
