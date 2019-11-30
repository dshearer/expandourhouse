FROM golang:1-alpine

WORKDIR /app/

# download 3rd-party libs
ADD src/go.mod ./src/
RUN cd src && go mod download

# compile app
ADD src/* ./src/
RUN cd src && go build

FROM alpine

WORKDIR /app
COPY --from=0 /app/src/loaddata loaddata
ADD data data
CMD ["/app/loaddata", "/app/data"]