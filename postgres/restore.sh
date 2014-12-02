#!/bin/bash
PGPASSFILE=/pgpass psql -U docker -h $POSTGRESADDRESS km < /backup/backup.sql
#PGPASSFILE=/pgpass psql -U docker -h $MAIN_PORT_5432_TCP_ADDR km
