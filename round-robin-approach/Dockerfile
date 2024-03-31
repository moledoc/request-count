FROM golang:1.22

WORKDIR /app
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o instance instance.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o entry entry.go
CMD ["mkdir" "/tmp"]
EXPOSE 8083
ENTRYPOINT ["/app/entry"]