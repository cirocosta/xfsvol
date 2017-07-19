FROM golang:alpine as builder

ADD ./main /go/src/github.com/cirocosta/xfsvol/main
ADD ./vendor /go/src/github.com/cirocosta/xfsvol/vendor
ADD ./manager /go/src/github.com/cirocosta/xfsvol/manager
ADD ./lib /go/src/github.com/cirocosta/xfsvol/lib

WORKDIR /go/src/github.com/cirocosta/xfsvol

RUN set -ex && \
  apk add --update gcc musl-dev make

RUN set -ex && \
  cd ./main && \
  go build -v  && \
  mv ./main /usr/bin/xfsvol

FROM busybox
COPY --from=builder /usr/bin/xfsvol /xfsvol

RUN mkdir -p /var/log/xfsvol /mnt/efs

CMD [ "xfsvol" ]
