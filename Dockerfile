FROM freekkalter/postgres-supervisord:km

ADD app /app
ADD supervisord.conf /etc/supervisor/conf.d/supervisord.conf


ADD pgpass /pgpass
RUN chmod 0600 /pgpass
ADD backup.sh /backup.sh
RUN chmod +x /backup.sh

EXPOSE 5432

CMD ["/usr/bin/supervisord"]
