FROM golang:latest

WORKDIR /app

# Pegando dependencias e arquivo de modulo
COPY go.mod ./

# Pegando o source file
COPY *.go ./

RUN go build -o tcpProxy .

EXPOSE 6380

CMD ["./tcpProxy"]
