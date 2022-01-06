FROM ubuntu:20.04
COPY ./web-monitor /
COPY ./config.conf /
COPY ./node.sh /
RUN chmod 777 /node.sh
RUN chmod 777 /web-monitor
RUN chmod 777 /config.conf
ENTRYPOINT ["/node.sh"]