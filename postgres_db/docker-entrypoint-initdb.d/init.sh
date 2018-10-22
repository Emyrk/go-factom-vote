#!/bin/bash
set -e

psql -f ../sql/database_init.sql
for filename in ../sql/functions/*.sql; do
     psql -f "$filename"
done

for filename in ../sql/data/*.sql; do
     psql -f "$filename"
done

for filename in ../sql/views/*.sql; do
     psql -f "$filename"
done


