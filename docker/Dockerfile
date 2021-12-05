FROM ubuntu:trusty 

WORKDIR /root

RUN apt-get update

# Install gtk headers
RUN apt-get install -y wget libgtk-3-dev libcairo2-dev libglib2.0-dev

# Install go
RUN wget -O go.tar.gz https://go.dev/dl/go1.17.4.linux-amd64.tar.gz; \
tar -xf go.tar.gz;

RUN apt-get install -y git build-essential clang pkg-config zip;

RUN git clone https://github.com/varnamproject/govarnam.git; \
git clone https://github.com/varnamproject/govarnam-ibus.git

ENV PATH="/root/go/bin:${PATH}"

WORKDIR /root/govarnam

RUN CC=clang make; \
sudo make install; \
make release

RUN mkdir -p /extract; \
cp *.zip /extract

WORKDIR /root/govarnam-ibus

RUN git checkout gtk-3.10-ubuntu-14.04; \
make ubuntu-14; \
make release

RUN cp *.zip /extract