FROM alpine:3.7

ADD dist/trigger /bin/trigger

CMD ["/bin/trigger"]