FROM golang:1.19.0-alpine3.16

RUN apk add  --no-cache ffmpeg
RUN apk add  --no-cache vorbis-tools

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /voice2fileBot


CMD [ "/voice2fileBot" ]
