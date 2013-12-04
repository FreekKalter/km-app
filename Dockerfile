FROM freekkalter/precise-vim
MAINTAINER freek@kalteronline.org

EXPOSE 5432
EXPOSE 4001

RUN apt-get update
# software-properties-common provides 'add-apt-repository' command
RUN apt-get install -y python-software-properties software-properties-common

RUN add-apt-repository ppa:pitti/postgresql
RUN echo "deb http://archive.ubuntu.com/ubuntu precise main universe" > /etc/apt/sources.list
RUN apt-get update

RUN apt-get -y install postgresql-9.2 postgresql-client-9.2 postgresql-contrib-9.2 sudo supervisor

RUN echo 'host all all 0.0.0.0/0 md5' >> /etc/postgresql/9.2/main/pg_hba.conf

# configure supervisord
RUN mkdir -p /var/log/supervisor
ADD supervisord.conf /etc/supervisor/conf.d/supervisord.conf
