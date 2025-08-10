FROM golang:1.23-alpine

WORKDIR /app

ENV GOPATH=/go
ENV PATH $PATH:/go/bin:$GOPATH/bin

COPY . /app

# build go app
RUN go mod download
RUN go build -o sse-server /app/cmd/main.go

CMD ["/app/sse-server"]
