FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/server ./cmd/main.go
RUN ls -la /app || dir /app
RUN ls -la /app/templates

EXPOSE 8080

CMD ["/app/server"]