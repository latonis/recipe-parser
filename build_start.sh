docker buildx build -t go-recipe-service ./service/
docker buildx build -t go-recipe-web ./web/

# docker run -p 9000:9000 go-recipe-service
# docker run -p 80:5500 go-recipe-web
