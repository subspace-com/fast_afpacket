# fast_afpacket

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
Bringing machine 'fast_afpacket-101' up with 'virtualbox' provider...
Bringing machine 'fast_afpacket-102' up with 'virtualbox' provider...
...
==> fast_afpacket-101: Machine 'fast_afpacket-101' has a post `vagrant up` message. This is a message
==> fast_afpacket-101: from the creator of the Vagrantfile, and not from Vagrant itself:
==> fast_afpacket-101: 
==> fast_afpacket-101: Vanilla Debian box. See https://app.vagrantup.com/debian for help and bug reports

==> fast_afpacket-102: Machine 'fast_afpacket-102' has a post `vagrant up` message. This is a message
==> fast_afpacket-102: from the creator of the Vagrantfile, and not from Vagrant itself:
==> fast_afpacket-102: 
==> fast_afpacket-102: Vanilla Debian box. See https://app.vagrantup.com/debian for help and bug reports

>  vagrant status
Current machine states:

fast_afpacket-101          running (virtualbox)
fast_afpacket-102          running (virtualbox)

This environment represents multiple VMs. The VMs are all listed
above with their current state. For more information about a specific
VM, run `vagrant status NAME`.
```

3. Open up a second terminal window in the root of the repo.

4. In the first terminal window SSH into the first virtual machine

```console
>  vagrant ssh fast_afpacket-101
Linux fast_afpacket-101 4.19.0-18-amd64 #1 SMP Debian 4.19.208-1 (2021-09-29) x86_64

The programs included with the Debian GNU/Linux system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
permitted by applicable law.
Last login: Wed May  4 17:12:12 2022 from 10.0.2.2
vagrant@fast_afpacket-101:~$
```

5. In the second terminal window SSH into the second virtual machine

```console
>  vagrant ssh fast_afpacket-102
Linux fast_afpacket-102 4.19.0-18-amd64 #1 SMP Debian 4.19.208-1 (2021-09-29) x86_64

The programs included with the Debian GNU/Linux system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
permitted by applicable law.
Last login: Wed May  4 17:12:12 2022 from 10.0.2.2
vagrant@fast_afpacket-102:~$
```

6. In each of the terminal windows change into the shared folder containing the
repo.

```console
vagrant@fast_afpacket-101:~$ cd /vagrant/
vagrant@fast_afpacket-101:/vagrant$
```

7. In the first terminal where you're connected `fast_afpacket-101` start the
example app to start it sending packets to the second instance which we will
start in the next step.

```console
vagrant@fast_afpacket-101:/vagrant$ ./run-example.sh

```

8. In the second terminal where you're connected `fast_afpacket-102` start the
example app to start it sending packets to the first instance.

```console
vagrant@fast_afpacket-102:/vagrant$ ./run-example.sh

```

9. Watching each terminal you will see timestamps being printed for each packet
being send and received.