FROM golang:1.17

RUN buildDeps=" \
        lsb-release \
    " \
    runtimeDeps=" \
        curl \
        git \
        vim \
        nano \
        htop \
    " \
    && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y $buildDeps $runtimeDeps

ENV GOPATH='/go-workspace'
ENV GOBIN=${GOPATH}/bin
ENV GOSRC=${GOPATH}/src
ENV PROJECTDIR=${GOSRC}/github.com/ashishkumar68/aws-ops

RUN mkdir -p ${GOPATH}/bin/
RUN mkdir -p ${GOPATH}/pkg/
RUN mkdir -p ${GOPATH}/src/

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s

EXPOSE 8080

COPY init/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

WORKDIR ${PROJECTDIR}
CMD ["air"]