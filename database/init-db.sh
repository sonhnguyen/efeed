#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER efeed;
    CREATE DATABASE efeed;
    GRANT ALL PRIVILEGES ON DATABASE efeed TO efeed;
EOSQL
