#! /bin/bash
apt-get update
wget https://go.dev/dl/go1.18.3.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.18.3.linux-amd64.tar.gz
rm go1.18.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /root/.profile
. ~/.profile
apt-get install -y git-core
mkdir /var/Go
cd /var/Go/ && git clone https://github.com/jhygt46/Final