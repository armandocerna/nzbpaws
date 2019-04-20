FROM golang:latest AS build-env
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor
ADD . /src
RUN cd /src && CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o nzbpaws

# final stage
FROM alpine:latest
WORKDIR /app
COPY --from=build-env /src/nzbpaws /app/
ENTRYPOINT /app/nzbpaws
