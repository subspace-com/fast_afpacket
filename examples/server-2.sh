go build -o fastafpacket fastafpacket.go

sudo ./fastafpacket \
    -iface-name=enp0s8 \
    -src-mac=02:42:ac:c8:01:62 \
    -src-addr=192.168.56.102 \
    -dst-mac=02:42:ac:c8:01:61 \
    -dst-addr=192.168.56.101