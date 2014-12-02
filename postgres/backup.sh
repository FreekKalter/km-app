#!/bin/bash
PGPASSFILE=/pgpass pg_dump --clean -U docker -h $POSTGRESADDRESS km > /backup/backup.sql
