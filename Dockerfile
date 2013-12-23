FROM freekkalter/postgres-supervisord:km


ADD app /app
ADD supervisord.conf /etc/supervisor/conf.d/supervisord.conf

CMD ["/usr/bin/supervisord"]
