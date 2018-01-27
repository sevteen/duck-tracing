FROM golang

RUN mkdir -p /go/src/github.com/sevteen/duck-tracing
WORKDIR /go/src/github.com/sevteen/duck-tracing

COPY . .

RUN go build -o main .

CMD ["./main"]