FROM golang:1.20

WORKDIR /gopritavegpt

RUN apt-get update && \
      apt-get upgrade -yq && \
      apt-get install -y build-essential make gcc g++ cmake

ADD go.mod go.sum .

RUN go mod tidy

ADD . .

RUN make build

ENTRYPOINT /goprivategpt/goprivategpt
