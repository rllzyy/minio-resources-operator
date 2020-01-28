FROM docker:19.03.5-dind as static-docker-source
FROM bitnami/kubectl:1.16.3-r54 as kubectl-source
FROM hadolint/hadolint:v1.17.4-2-g48db165-alpine as hadolint-source
FROM minio/mc:RELEASE.2020-01-13T22-49-03Z as mc-source
FROM golang:1.13.6-alpine3.11

COPY --from=static-docker-source /usr/local/bin/docker /usr/local/bin/docker
COPY --from=hadolint-source /bin/hadolint /usr/local/bin/hadolint
COPY --from=kubectl-source /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
COPY --from=mc-source /usr/bin/mc /usr/local/bin/mc

ENV WATCH_NAMESPACE ""
ENV OPERATOR_NAME "minio-resources-operator"

RUN apk add --no-cache \
      vim \
      htop \
      gcc \
      musl-dev \
      openssh-client \
      linux-headers \
      git && \
    go get -v golang.org/x/tools/gopls && \
    go get -u -v github.com/ramya-rao-a/go-outline && \
    go get -u -v github.com/acroca/go-symbols && \
    go get -u -v github.com/mdempsky/gocode && \
    go get -u -v github.com/rogpeppe/godef && \
    go get -u -v golang.org/x/tools/cmd/godoc && \
    go get -u -v github.com/zmb3/gogetdoc && \
    go get -u -v golang.org/x/lint/golint && \
    go get -u -v github.com/fatih/gomodifytags && \
    go get -u -v golang.org/x/tools/cmd/gorename && \
    go get -u -v sourcegraph.com/sqs/goreturns && \
    go get -u -v golang.org/x/tools/cmd/goimports && \
    go get -u -v github.com/cweill/gotests/... && \
    go get -u -v golang.org/x/tools/cmd/guru && \
    go get -u -v github.com/josharian/impl && \
    go get -u -v github.com/haya14busa/goplay/cmd/goplay && \
    go get -u -v github.com/uudashr/gopkgs/cmd/gopkgs && \
    go get -u -v github.com/davidrjenni/reftools/cmd/fillstruct && \
    go get -u -v github.com/alecthomas/gometalinter && \
    go get -u -v github.com/stamblerre/gocode && \
    go get -u -v golang.org/x/tools/cmd/stringer && \
    mv /go/bin/gocode /go/bin/gocode-gomod && \
    gometalinter --install && \
    go get -u -v github.com/go-delve/delve/cmd/dlv && \
    rm -rf \
      /root/.cache/go-mod/ \
      /go/src/ \
      /go/pkg/ && \
    wget -O /usr/local/bin/operator-sdk \
      https://github.com/operator-framework/operator-sdk/releases/download/v0.14.1/operator-sdk-v0.14.1-x86_64-linux-gnu && \
    chmod 755 /usr/local/bin/operator-sdk && \
    mkdir -p /root/.kube