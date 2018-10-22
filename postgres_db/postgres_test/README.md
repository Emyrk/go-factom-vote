https://hub.docker.com/_/postgres/

`docker run --name some-app --link some-postgres:postgres -d application-that-uses-postgres`

`EXPOSE 5432`