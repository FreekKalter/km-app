#!/bin/bash
sudo chown -R postgres:postgres /data
sudo -u postgres /usr/lib/postgresql/9.2/bin/postgres -D /data/main -c config_file=/data/postgresql.conf -c listen_addresses=*
