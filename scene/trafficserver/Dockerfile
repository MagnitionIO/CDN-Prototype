FROM ubuntu:23.10

RUN apt-get update \
    && apt-get install -y curl trafficserver \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /run/trafficserver \
    && chmod 777 /run/trafficserver

RUN mkdir -p /var/cache

USER trafficserver
WORKDIR /home/trafficserver

CMD ["traffic_server"]

# FROM debian:buster

# ARG VERSION=9.2.2

# RUN DEBIAN_FRONTEND=noninteractive apt-get update \
#         && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y \
#         && DEBIAN_FRONTEND=noninteractive apt-get install -y curl locales build-essential bzip2 libssl-dev libxml2-dev libpcre3-dev tcl-dev libboost-dev \
#         && rm -rf /var/lib/apt/lists/* \
#         && export LANGUAGE=en_US.UTF-8 LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8 \
#         && echo "LC_ALL=en_US.UTF-8" >> /etc/environment && echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen && echo "LANG=en_US.UTF-8" > /etc/locale.conf \
#         && locale-gen en_US.UTF-8 \
#         && DEBIAN_FRONTEND=noninteractive dpkg-reconfigure locales \
#         && mkdir -p /tmp/trafficserver \
#         && curl -L https://downloads.apache.org/trafficserver/trafficserver-${VERSION}.tar.bz2 | tar xjvf - -C /tmp/trafficserver --strip-components 1 \
#         && cd /tmp/trafficserver \
#         && ./configure \
#         && make \
#         && make install \
#         && rm -rf /tmp/trafficserver

# EXPOSE 8080

# CMD ["traffic_server"]
