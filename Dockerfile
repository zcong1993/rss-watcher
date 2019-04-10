FROM golang:1.12.1 AS build
WORKDIR /mnt
COPY . .
RUN CGO_ENABLED=0 go build -o ./bin/rw main.go

FROM alpine:3.7
WORKDIR /opt
EXPOSE 1234
EXPOSE 8080
RUN apk add --no-cache ca-certificates
COPY --from=build /mnt/bin/* /usr/bin/
CMD ["rw"]
