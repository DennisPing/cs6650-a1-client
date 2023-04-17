# Twinder Client

## Getting Started

### Build
```
make
```

### Run
```
./client
```

## Watch out

The Swagger API specifies the POST endpoint as:
```
/swipe/{leftorright}/
```

The trailing slash must be there! This is different than:
```
/swipe/{leftorright}
```

## Build and Test Docker image
```
docker build -t client:a1 .
```

## Create Docker network
```
docker network create my_network
```

```
docker run --name client --network my_network -e SERVER_URL=http://server:8080 -e LOG_LEVEL=Warn client:a1
```

## Stop local Docker container
```
docker ps
docker stop <container id>
docker network rm my_network
```

## Tag and push Docker image to Google Container Registry
```
docker tag client:a1 gcr.io/cs6650-dping/client:a1
docker push gcr.io/cs6650-dping/client:a1
```

## Deploy new Cloud Run service
```
gcloud run deploy client \
    --image gcr.io/cs6650-dping/client:a1 \
    --platform managed \
    --region us-central1 \
    --allow-unauthenticated \
    --set-env-vars="LOG_LEVEL=Info" \
    --set-env-vars="SERVER_URL=https://server-[random]-uc.a.run.app"
```
