FROM golang:alpine as builder

ADD ./main /go/src/github.com/cirocosta/xfsvol/main
ADD ./vendor /go/src/github.com/cirocosta/xfsvol/vendor
ADD ./manager /go/src/github.com/cirocosta/xfsvol/manager

WORKDIR /go/src/github.com/cirocosta/xfsvol
RUN set -ex && \
  cd ./main && \
  CGO_ENABLED=0 go build -v -a -ldflags '-extldflags "-static"' && \
  mv ./main /usr/bin/xfsvol

FROM busybox
COPY --from=builder /usr/bin/xfsvol /xfsvol

RUN mkdir -p /var/log/xfsvol /mnt/efs

CMD [ "xfsvol" ]
