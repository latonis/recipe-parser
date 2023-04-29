if docker ps --format '{{.ID}}'; then
    docker stop `docker ps --format '{{.ID}}'`
fi
docker buildx build -t go-recipe .
docker run -p 9000:9000 go-recipe 