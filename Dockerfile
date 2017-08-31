FROM golang as builder

RUN set -ex && \
  apt update -y && \
  apt install -y xfsprogs build-essential

ADD ./main /go/src/github.com/cirocosta/xfsvol/main
ADD ./VERSION /go/src/github.com/cirocosta/xfsvol/main/VERSION
ADD ./vendor /go/src/github.com/cirocosta/xfsvol/vendor
ADD ./manager /go/src/github.com/cirocosta/xfsvol/manager
ADD ./lib /go/src/github.com/cirocosta/xfsvol/lib

WORKDIR /go/src/github.com/cirocosta/xfsvol

RUN set -ex && \
  cd ./main && \
  go build \
        -tags netgo -v -a \
        -ldflags "-X main.version=$(cat ./VERSION) -extldflags \"-static\"" && \
  mv ./main /usr/bin/xfsvol


FROM busybox
COPY --from=builder /usr/bin/xfsvol /xfsvol

RUN mkdir -p /var/log/xfsvol /mnt/efs

CMD [ "xfsvol" ]
