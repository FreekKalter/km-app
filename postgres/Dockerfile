FROM ubuntu:13.10
MAINTAINER freek@kalteronline.org

EXPOSE 5432

# Update OS.
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy main" > /etc/apt/sources.list
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy-updates main" >> /etc/apt/sources.list
RUN echo "deb http://security.ubuntu.com/ubuntu saucy-security main universe" >> /etc/apt/sources.list
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy universe" >> /etc/apt/sources.list
RUN echo "deb http://archive.ubuntu.com/ubuntu saucy-updates universe" >> /etc/apt/sources.list
RUN apt-get update

## software-properties-common provides 'add-apt-repository' command
#RUN apt-get install -y python-software-properties software-properties-common
#
#RUN add-apt-repository ppa:pitti/postgresql
#RUN apt-get update

RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8
RUN echo "deb http://apt.postgresql.org/pub/repos/apt/ saucy-pgdg main" >> /etc/apt/sources.list.d/pgdg.list

RUN apt-get update
RUN apt-get upgrade -y

RUN apt-get -y install postgresql-9.2 postgresql-client-9.2 postgresql-contrib-9.2 sudo supervisor

RUN echo 'host all all 0.0.0.0/0 md5' >> /etc/postgresql/9.2/main/pg_hba.conf

ADD pgpass /pgpass
RUN chmod 0600 /pgpass

ADD backup.sh /backup.sh
RUN chmod +x /backup.sh

ADD restore.sh /restore.sh
RUN chmod +x /restore.sh

ADD start_postgres.sh /start_postgres.sh
RUN chmod +x /start_postgres.sh

CMD /start_postgres.sh
