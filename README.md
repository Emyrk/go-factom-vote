
# Go-Factom-Vote

This is the Factom blockchain voting daemon. It parses the blockchain 
to monitor ongoing votes, stores them in an postgres database, and exposes
a quick api to interact with voting related data.

The daemon composes of a few parts:
- Factomd
    - Factomd is required, as the data is all in the blockchain.
- The postgres instance
    - The postgres instance has a schema for vote related objects. This
    instance needs to be present.
- The scraper
    - The scraper will pull data from factomd, validate the data, and insert it
    into the postgres instance.
- The apiserver
    - The apiserver exposes a graphql api that allows for queries into the postgres
    instance.
    
Each of these is a docker container represented in the `docker-compose.yml`.
Because the scraper and apiserver are seperated, the reading and writing instances
can be scaled/managed separately. 

# Voting Spec

https://docs.google.com/document/d/137gw8JTqKZdfe-AF0e02pFgyLsPCTjYbX0mguC_N3GE/edit?ts=5b71a406#


# Running your own daemon

The easiest way to run your own daemon, is to use the docker-compose. This will spin
up all the instances described above. Edit the `factomd.env` and `factomd.conf` prior to doing this
if you do not wish to run on testnet.

```
docker-compose up -d
```

It is possible to use your own factomd, but editing of the dockerfiles
is required.


# Individual container update

```
docker-compose up -d --no-deps --build scraper
```

# API

The api uses graphql, and documentation can be found in the playground at `localhost/graphql`

