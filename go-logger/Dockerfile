FROM golang:1.19

WORKDIR /app

# Pegando dependencias e arquivo de modulo
COPY go.mod ./

# Pegando o source file
COPY *.go ./

RUN go build -o ./tcp-proxy

EXPOSE 6380

CMD ["./tcp-proxy"]
