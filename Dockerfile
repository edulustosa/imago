FROM golang:1.23.5-alpine

RUN apk add --no-cache \
    libwebp-dev \
    build-base \
    gcc \
    musl-dev \
    pkgconfig

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./bin/imago .

EXPOSE 8080

CMD [ "./bin/imago" ]
