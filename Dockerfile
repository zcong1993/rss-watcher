FROM golang:1.16 AS build
WORKDIR /mnt
ADD go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/rw ./cmd/main.go

FROM alpine:3
WORKDIR /opt
RUN apk add --no-cache ca-certificates
COPY --from=build /mnt/bin/* /usr/bin/
CMD ["rw"]
