FROM alpine:latest
RUN apk add tzdata
COPY main /
COPY ./settings.ini /
ENTRYPOINT /main
