#!/bin/bash
PGPASSFILE=/pgpass psql --clean -U docker -h $MAIN_PORT_5432_TCP_ADDR km < /backup/backup.sql
