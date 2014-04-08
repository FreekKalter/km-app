FROM freekkalter/nginx:start_nginx

RUN apt-get update && apt-get upgrade
ADD default.template /etc/nginx/sites-available/default.template
ADD static /static
