FROM golang as builder

RUN set -ex && \
        apt update -y && \
        apt install -y \
                xfsprogs \
                build-essential \
                xfslibs-dev

ADD ./ /go/src/github.com/cirocosta/xfsvol/

WORKDIR /go/src/github.com/cirocosta/xfsvol

RUN set -ex && \
        cd ./plugin && \
        go build \
                -tags netgo -v -a \
                -ldflags "-X main.version=$(cat ./VERSION) -extldflags \"-static\"" && \
        mv ./plugin /usr/bin/xfsvol


FROM busybox
COPY --from=builder /usr/bin/xfsvol /xfsvol

CMD [ "/xfsvol" ]
