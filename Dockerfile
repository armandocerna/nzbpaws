FROM golang:latest AS build-env
ENV GO111MODULE=on
ADD . /src
RUN cd /src && go build -o nzbpaws

# final stage
FROM alpine:latest
WORKDIR /app
COPY --from=build-env /src/nzbpaws /app/
ENTRYPOINT ./nzbpaws