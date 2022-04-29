FROM alpine:3.10

# Install some needed packages
RUN apk add --no-cache\
      netcat-openbsd \
      libc6-compat \
      ca-certificates \
      tzdata

COPY ./aws_appconfig /aws_appconfig
CMD ./aws_appconfig

