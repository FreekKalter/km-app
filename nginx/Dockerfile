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

RUN apt-get update
RUN apt-get upgrade -y

# Install basic packages.
RUN apt-get install -y software-properties-common
RUN apt-get install -y curl git htop unzip vim wget


# Install Nginx.
RUN add-apt-repository -y ppa:nginx/stable
RUN apt-get update
RUN apt-get install -y nginx

EXPOSE 443

ADD start_nginx /start_nginx
RUN chmod +x /start_nginx

ADD default.template /etc/nginx/sites-available/default.template
ADD nginx.template /etc/nginx/nginx.template
ADD client_auth.conf /etc/nginx/client_auth.conf
ADD static /static
