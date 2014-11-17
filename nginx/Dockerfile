#
# Nginx Dockerfile
#
# https://github.com/dockerfile/nginx
#
FROM ubuntu:13.10

# Update OS.
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy main" > /etc/apt/sources.list
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy-updates main" >> /etc/apt/sources.list
RUN echo "deb http://security.ubuntu.com/ubuntu saucy-security main universe" >> /etc/apt/sources.list
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy universe" >> /etc/apt/sources.list
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy-updates universe" >> /etc/apt/sources.list

# install add-apt-repository to add the nginx repo
RUN apt-get update && apt-get install -y curl software-properties-common

# Install Nginx.
RUN add-apt-repository -y ppa:nginx/stable
RUN apt-get update && apt-get install -y nginx

EXPOSE 443

# Move all configs in the right locations
RUN rm /etc/nginx/nginx.conf
ADD default /etc/nginx/sites-available/default
ADD client_auth.conf /etc/nginx/client_auth.conf
ADD static /static

# install confd
WORKDIR /usr/local/bin
RUN curl -L https://github.com/kelseyhightower/confd/releases/download/v0.5.0/confd-0.5.0-linux-amd64 -o confd
RUN chmod +x confd

# configure confd
RUN ["/bin/bash", "-c", "mkdir -p /etc/confd/{conf.d,templates}"]
ADD nginx.conf.tmpl /etc/confd/templates/nginx.conf.tmpl
ADD nginx.toml /etc/confd/conf.d/nginx.toml
ADD confd-watch /usr/local/bin/confd-watch
RUN chmod +x /usr/local/bin/confd-watch

CMD /usr/local/bin/confd-watch
