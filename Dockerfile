FROM alpine

ADD bifrost /bifrost
ADD config/config.json config/config.json

ENTRYPOINT [ "/bifrost" ]