# -*- mode: ruby -*-
# # vi: set ft=ruby :

Vagrant.configure("2") do |config|

    # reload for installing packages and rebooting
    # vbguest for mounting shared folders into the vm
    config.vagrant.plugins = ["vagrant-reload", "vagrant-vbguest"]

    # use VirtualBox provider and provision CPU and RAM resources
    config.vm.provider :virtualbox do |vboxconfig|
        vboxconfig.memory = 8192
        vboxconfig.cpus = 2
    end

    # provision each vm via the script
    config.vm.provision :shell, path: "vagrant-bootstrap.sh"
    config.vm.provision :reload

    # define the first vm and set up a private network
    # with preconfigured IP, MAC, and ARP table entries of the
    # second vm below
    config.vm.define "fast-afpacket-101" do |boxconfig|
        boxconfig.vm.box = "debian/buster64"
        
        boxconfig.vm.hostname = "fast-afpacket-101"
        
        boxconfig.vm.network "private_network", ip: "192.168.56.101", :mac => "0242acc80161"
        boxconfig.vm.provision :shell, run: "always", inline: "ip address replace fde4:8dba:82e1::c1/64 dev eth1"

        boxconfig.vm.provision :shell, run: "always", inline: "ip neigh replace 192.168.56.102 lladdr 02:42:ac:c8:01:62 dev eth1 nud reachable"
        boxconfig.vm.provision :shell, run: "always", inline: "ip neigh replace fde4:8dba:82e1::c2 lladdr 02:42:ac:c8:01:62 dev eth1 nud reachable"
    end

    # define the second vm in the same private network
    # with preconfigured IP, MAC, and ARP table entries of the
    # first vm above
    config.vm.define "fast-afpacket-102" do |boxconfig|
        boxconfig.vm.box = "debian/buster64"
        
        boxconfig.vm.hostname = "fast-afpacket-102"
        
        boxconfig.vm.network "private_network", ip: "192.168.56.102", :mac => "0242acc80162"
        boxconfig.vm.provision :shell, run: "always", inline: "ip address replace fde4:8dba:82e1::c2/64 dev eth1"

        boxconfig.vm.provision :shell, run: "always", inline: "ip neigh replace 192.168.56.101 lladdr 02:42:ac:c8:01:61 dev eth1 nud reachable"
        boxconfig.vm.provision :shell, run: "always", inline: "ip neigh replace fde4:8dba:82e1::c1 lladdr 02:42:ac:c8:01:61 dev eth1 nud reachable"
    end

end