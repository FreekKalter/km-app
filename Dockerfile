FROM freekkalter/postgres-supervisord:km


ADD app /app
ADD supervisord.conf /etc/supervisor/conf.d/supervisord.conf

EXPOSE 5432

CMD ["/usr/bin/supervisord"]
