#!/bin/bash
PGPASSFILE=/pgpass pg_dump -U docker -h $MAIN_PORT_5432_TCP_ADDR km > /backup/backup.sql
