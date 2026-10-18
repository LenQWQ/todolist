FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY build/app-linux ./app
COPY web ./web
COPY etc/todo-api.docker.yaml ./etc/todo-api.yaml
EXPOSE 8888
CMD ["./app"]