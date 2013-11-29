FROM ubuntu

RUN apt-get update
RUN apt-get install -y python-software-properties software-properties-common
RUN add-apt-repository ppa:pitti/postgresql
RUN apt-get update

RUN apt-get -y install postgresql-9.2 postgresql-client-9.2 postgresql-contrib-9.2 sudo

RUN echo 'host all all 0.0.0.0/0 md5' >> /etc/postgresql/9.2/main/pg_hba.conf

ADD start /
RUN chmod +x /start
CMD /start
