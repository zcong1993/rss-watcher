FROM golang:1.15 AS build
WORKDIR /mnt
ADD go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/rw ./cmd/main.go

FROM alpine:3
WORKDIR /opt
EXPOSE 1234
EXPOSE 8080
RUN apk add --no-cache ca-certificates
COPY --from=build /mnt/bin/* /usr/bin/
CMD ["rw"]
