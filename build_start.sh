if docker ps --format '{{.ID}}'; then
    docker stop `docker ps --format '{{.ID}}'`
fi
docker buildx build -t go-recipe-service ./service/
# docker buildx build -t go-recipe-web ./web/
docker run -p 9000:9000 go-recipe-service
# docker run -p 5500:80 go-recipe-web