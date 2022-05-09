# fast_afpacket

fast_afpacket is a Go library for sending and receiving packets using
[AF_PACKET](https://man7.org/linux/man-pages/man7/packet.7.html) sockets. It's
built around [mdlayher/socket](https://github.com/mdlayher/socket). The library
also fully supports
[linux socket timestamp options](https://www.kernel.org/doc/html/latest/networking/timestamping.html),
setting the correct ioctls and socket options automatically, and includes APIs
for retrieving RX/TX timestamps back from the socket.

## Requirements

Because this library deals with interacting with raw packets and directly with
network interfaces you need the following:

- `libpcap-dev` installed for raw packet encoding/decoding
- `sudo` to run with root privileges so you program can access the network interfaces

## Example App

For convenience there is an example service which acts as a client and server to
send packets to other instances of itself that is running.

To set up the network in which to get the example working we use Vagrant to
create compete Linux virtual machines and specifically configured network
interfaces and their associated MAC and IPv4/6 addresses and ARP table entries.

### Running

1. Install [Vagrant](https://www.vagrantup.com) for your operating system.

2. From the root folder where the `Vagrantfile` is located run `vagrant up`. You should now have 2 virtual machines running:

```console
>  vagrant up
Bringing machine 'fast-afpacket-101' up with 'virtualbox' provider...
Bringing machine 'fast-afpacket-102' up with 'virtualbox' provider...
...
==> fast-afpacket-101: Machine 'fast-afpacket-101' has a post `vagrant up` message. This is a message
==> fast-afpacket-101: from the creator of the Vagrantfile, and not from Vagrant itself:
==> fast-afpacket-101:
==> fast-afpacket-101: Vanilla Debian box. See https://app.vagrantup.com/debian for help and bug reports

==> fast-afpacket-102: Machine 'fast-afpacket-102' has a post `vagrant up` message. This is a message
==> fast-afpacket-102: from the creator of the Vagrantfile, and not from Vagrant itself:
==> fast-afpacket-102:
==> fast-afpacket-102: Vanilla Debian box. See https://app.vagrantup.com/debian for help and bug reports

>  vagrant status
Current machine states:

fast-afpacket-101          running (virtualbox)
fast-afpacket-102          running (virtualbox)

This environment represents multiple VMs. The VMs are all listed
above with their current state. For more information about a specific
VM, run `vagrant status NAME`.
```

3. Open up a second terminal window in the root of the repo.

4. In the first terminal window SSH into the first virtual machine

```console
>  vagrant ssh fast-afpacket-101
...
Last login: Wed May  4 17:12:12 2022 from 10.0.2.2
vagrant@fast-afpacket-101:~$
```

5. In the second terminal window SSH into the second virtual machine

```console
>  vagrant ssh fast-afpacket-102
...
Last login: Wed May  4 17:12:12 2022 from 10.0.2.2
vagrant@fast-afpacket-102:~$
```

6. In each of the terminal windows change into the shared folder containing the
repo.

```console
vagrant@fast-afpacket-101:~$ cd /vagrant/
vagrant@fast-afpacket-101:/vagrant$
```

7. In the first terminal where you're connected `fast-afpacket-101` start the
example app to start it sending packets to the second instance which we will
start in the next step.

```console
vagrant@fast-afpacket-101:/vagrant$ ./server-1.sh

```

8. In the second terminal where you're connected `fast-afpacket-102` start the
example app to start it sending packets to the first instance.

```console
vagrant@fast-afpacket-102:/vagrant$ ./server-2.sh

```

9. Watching each terminal you will see timestamps being printed for each packet
being send and received.

**fast-afpacket-101**
```console
INFO[0013] TX Recvmsg                                    hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=13 software="2022-05-09T23:16:42.507654453Z" software_ns=1652138202507654453
INFO[0013] RX Recvmsg                                    delay="-423.796µs" hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=33 software="2022-05-09T23:16:42.855011458Z" software_ns=1652138202855011458 userspace="2022-05-09T23:16:42.855435254Z" userspace_ns=1652138202855435254
INFO[0014] TX Recvmsg                                    hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=14 software="2022-05-09T23:16:43.50814104Z" software_ns=1652138203508141040
INFO[0014] RX Recvmsg                                    delay="-393.518µs" hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=34 software="2022-05-09T23:16:43.85478227Z" software_ns=1652138203854782270 userspace="2022-05-09T23:16:43.855175788Z" userspace_ns=165213820385517578
```

**fast-afpacket-102**
```console
INFO[0034] TX Recvmsg                                    hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=34 software="2022-05-09T23:16:43.853715657Z" software_ns=1652138203853715657
INFO[0034] RX Recvmsg                                    delay="-324.866µs" hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=15 software="2022-05-09T23:16:44.508531621Z" software_ns=1652138204508531621 userspace="2022-05-09T23:16:44.508856487Z" userspace_ns=1652138204508856487
INFO[0035] TX Recvmsg                                    hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=35 software="2022-05-09T23:16:44.853216023Z" software_ns=1652138204853216023
INFO[0035] RX Recvmsg                                    delay="-448.397µs" hardware="0001-01-01T00:00:00Z" hardware_ns=-6795364578871345152 probe=16 software="2022-05-09T23:16:45.507733046Z" software_ns=1652138205507733046 userspace="2022-05-09T23:16:45.508181443Z" userspace_ns=1652138205508181443
```

Note: Because we are using virtual machines to run the example you will not see
any hardware timestamps. You will only see software timestamps which is the
kernel setting the timestamp on the packet. In order to get real hardware
timestamps you will need to use the library on bare metal hardware with NICs which
support hardware timestamping.


### Authors

fast_afpacket was designed and authored by [Blain Smith](https://github.com/blainsmith) and [Joe williams](https://github.com/joewilliams) at [Subspace](https://subspace.com/).
