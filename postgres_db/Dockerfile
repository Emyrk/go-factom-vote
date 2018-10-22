FROM postgres:10.5
COPY docker-entrypoint-initdb.d /docker-entrypoint-initdb.d
COPY sql /sql

RUN apt-get update \
    && apt-get -y install curl vim \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

EXPOSE 5432

# We need to initialize the database