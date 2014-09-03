FROM ubuntu:14.04

RUN apt-get update
RUN apt-get install apache2 libapache2-mod-php5 php5 php5-pgsql -y

ADD 001-www.conf /etc/apache2/sites-available/

RUN a2enmod rewrite
RUN a2ensite 001-www

ENV APACHE_RUN_USER www-data
ENV APACHE_RUN_GROUP www-data
ENV APACHE_LOG_DIR /var/log/apache2
ENV APACHE_PID_FILE /var/run/apache2.pid
ENV APACHE_RUN_DIR /var/run/apache2
ENV APACHE_LOCK_DIR /var/lock/apache2
ENV APACHE_SERVERADMIN admin@localhost
ENV APACHE_SERVERNAME localhost
ENV APACHE_SERVERALIAS docker.localhost
ENV APACHE_DOCUMENTROOT /var/www

RUN rm /var/www/html/*

ADD www /var/www/html

VOLUME /var/www/html

EXPOSE 80

CMD ["/usr/sbin/apache2", "-D", "FOREGROUND"]
