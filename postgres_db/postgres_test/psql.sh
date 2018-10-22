echo 'password is "password"'
docker run -it --rm --link some-postgres:postgres postgres psql -h postgres -U postgres
psql (9.5.0)