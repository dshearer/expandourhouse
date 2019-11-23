FROM golang:1-alpine

WORKDIR /api

# download 3rd-party libs
ADD src/go.mod ./src/
RUN cd src && go mod download

# compile app
ADD src/* ./src/
RUN cd src && go build

FROM alpine

COPY --from=0 /api/src/api /api
CMD ["/api"]