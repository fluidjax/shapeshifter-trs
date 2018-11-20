FROM ubuntu:latest

# Install Basics
RUN apt-get update
RUN apt-get --yes install build-essential git
RUN apt-get --yes install curl
RUN apt-get --yes install python-setuptools python-dev build-essential
RUN apt-get --yes install python-pip
RUN apt-get --yes install cmake
RUN apt-get --yes install libtool
RUN apt-get --yes install autoconf

# Install Go
WORKDIR /usr/local/
RUN curl -O https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.11.2.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV LD_LIBRARY_PATH=/usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib:$LD_LIBRARY_PATH
WORKDIR /root/go/src

COPY . .

# Get Go Deps
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/aws/aws-sdk-go/aws
RUN go get github.com/google/uuid

# Install AWS SDK - For DynamoDB
RUN pip install awscli --upgrade --user
RUN mv .aws ~/

# Install Crypto Library
ARG token
WORKDIR /usr/local/
RUN token=$(cat /root/go/src/.gitToken) && git clone https://$token@github.com/qredo/Qredo-Crypto-Library.git
WORKDIR /usr/local/Qredo-Crypto-Library/
RUN git checkout develop
WORKDIR /usr/local/Qredo-Crypto-Library/qredolib
RUN mkdir build
WORKDIR /usr/local/Qredo-Crypto-Library/qredolib/build
RUN cmake -DCMAKE_INSTALL_PREFIX=./binaries ..
RUN make install
RUN cp /usr/local/Qredo-Crypto-Library/qredolib/build/binaries/bin/qredo_test /usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib
WORKDIR /usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib
RUN LD_LIBRARY_PATH=. ./qredo_test

EXPOSE 5000

WORKDIR /root/go/src
CMD go run *.go





