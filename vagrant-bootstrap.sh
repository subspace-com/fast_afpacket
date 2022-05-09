#!/bin/bash

apt-get update

apt-get install -y linux-base
apt-get install -y libpcap-dev

curl -OLs https://golang.org/dl/go1.17.9.linux-amd64.tar.gz
tar -C /usr/local -xvf go1.17.9.linux-amd64.tar.gz

echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile

source ~/.profile
