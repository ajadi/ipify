FROM golang:1.21-alpine AS build
RUN apk add --no-cache git
ENV GO111MODULE=off
ENV GOPATH=/go
ENV CGO_ENABLED=0
RUN mkdir -p /go/src/github.com/rdegges && \
    git clone --depth 1 https://github.com/rdegges/ipify-api.git /go/src/github.com/rdegges/ipify-api
COPY get_ip.go /go/src/github.com/rdegges/ipify-api/api/get_ip.go
WORKDIR /go/src/github.com/rdegges/ipify-api
RUN go build -o /ipify-api .

FROM alpine:3.19
RUN adduser -D -u 1000 ipify
USER ipify
COPY --from=build /ipify-api /usr/local/bin/ipify-api
ENV PORT=3000
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/ipify-api"]
