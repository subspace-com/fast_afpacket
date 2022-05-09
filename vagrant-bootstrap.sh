#!/bin/bash

apt-get update

apt-get install -y libpcap-dev

curl -OLs https://golang.org/dl/go1.17.9.linux-amd64.tar.gz
tar -C /usr/local -xf go1.17.9.linux-amd64.tar.gz

echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/vagrant/.profile
