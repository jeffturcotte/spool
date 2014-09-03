FROM ubuntu:14.04

RUN apt-get update
RUN apt-get install -y openssh-server

RUN mkdir /var/run/sshd

RUN echo 'root:test' | chpasswd

RUN sed -i -e"s/^PermitRootLogin without-password/PermitRootLogin yes/" /etc/ssh/sshd_config

EXPOSE 22

CMD ["/usr/sbin/sshd", "-D"]
