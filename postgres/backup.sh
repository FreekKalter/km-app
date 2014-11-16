#!/bin/bash
PGPASSFILE=/pgpass pg_dump --clean -U docker -h $MAIN_PORT_5432_TCP_ADDR km > /backup/backup.sql
