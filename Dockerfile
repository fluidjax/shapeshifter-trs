FROM ubuntu:latest

# Install Crypto Library

ARG token
RUN apt-get update
RUN apt-get --yes install build-essential git
WORKDIR /usr/local/
RUN git clone https://$token@github.com/qredo/Qredo-Crypto-Library.git
WORKDIR /usr/local/Qredo-Crypto-Library/
RUN git checkout develop
WORKDIR /usr/local/Qredo-Crypto-Library/qredolib
RUN mkdir build
WORKDIR /usr/local/Qredo-Crypto-Library/qredolib/build
RUN apt-get --yes install cmake
RUN apt-get --yes install libtool
RUN apt-get --yes install autoconf
RUN cmake -DCMAKE_INSTALL_PREFIX=./binaries ..
RUN make install
RUN cp /usr/local/Qredo-Crypto-Library/qredolib/build/binaries/bin/qredo_test /usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib
WORKDIR /usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib
RUN LD_LIBRARY_PATH=. ./qredo_test

# Install Go

RUN apt-get --yes install curl
WORKDIR /usr/local/
RUN ls
RUN curl -O https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.11.2.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
WORKDIR /go/src/
ENV LD_LIBRARY_PATH=/usr/local/Qredo-Crypto-Library/qredolib/build/binaries/lib:$LD_LIBRARY_PATH
COPY . .
CMD go run trs.go





