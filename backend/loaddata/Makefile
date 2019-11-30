DOCKER_IMAGE = loaddata
DOCKER_OPTS = -it

include ../local-dev/local-dev.mk

.PHONY: get-data
get-data: \
	data/CVAP_2013-2017_ACS_csv_files.zip \
	data/CVAP_2012-2016_ACS_csv_files.zip

data/CVAP_2013-2017_ACS_csv_files.zip:
	mkdir -p data
	wget -P data https://www2.census.gov/programs-surveys/decennial/rdo/datasets/2017/2017-cvap/CVAP_2013-2017_ACS_csv_files.zip

data/CVAP_2012-2016_ACS_csv_files.zip:
	mkdir -p data
	wget -P data https://www2.census.gov/programs-surveys/decennial/rdo/datasets/2016/2016-cvap/CVAP_2012-2016_ACS_csv_files.zip