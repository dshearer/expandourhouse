FROM golang:1-alpine

WORKDIR /cvap-updater/

# download 3rd-party libs
ADD src/go.mod ./src/
RUN cd src && go mod download

# compile app
ADD src/* ./src/
RUN cd src && go build

FROM alpine

COPY --from=0 /cvap-updater/src/cvap-updater /cvap-updater
ADD CVAP_2013-2017_ACS_csv_files.zip /
ADD states.json /
CMD ["/cvap-updater", "-datazip", "/CVAP_2013-2017_ACS_csv_files.zip", "/states.json"]