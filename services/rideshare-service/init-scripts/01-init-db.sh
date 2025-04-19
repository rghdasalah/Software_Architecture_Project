#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "postgres" <<-EOSQL
  CREATE DATABASE rideshare;
EOSQL

psql -v ON_ERROR_STOP=1 --username "postgres" --dbname "rideshare" -f /docker-entrypoint-initdb.d/init_schema.pgsql