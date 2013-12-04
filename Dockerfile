FROM freekkalter/postgres-supervisord:km


ADD app /app
CMD ["/usr/bin/supervisord"]
