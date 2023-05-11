FROM alpine:latest
LABEL maintainer="OIT AV Services <oitav@byu.edu>"

ARG NAME

COPY ${NAME} /app
COPY website /website

ENTRYPOINT ["/app"]
