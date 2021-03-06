# -*- mode: ruby -*-
# vim: set ft=ruby :

VAGRANTFILE_API_VERSION = "2"

$provision = <<SCRIPT
# install deps
apt-get update; apt-get install -y git golang

# install latest docker
curl -sSL https://get.docker.io/ubuntu/ | sudo sh

# add vagrant user to docker group
gpasswd -a vagrant docker

# configure env
echo "export GOPATH=/home/vagrant/go" >> /home/vagrant/.bash_aliases
echo 'export PATH="$GOPATH/bin:$PATH"' >> /home/vagrant/.bash_aliases

# link spool dir for easy access
ln -s /home/vagrant/go/src/github.com/jeffturcotte/spool /home/vagrant/spool
SCRIPT

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  # See the online documentation at docs.vagrantup.com.

  # Every Vagrant virtual environment requires a box to build off of.
  config.vm.box = "ubuntu/trusty64"

  # Provision with bootstrap file.
  config.vm.provision :shell, inline: $provision

  # Create a private network, which allows host-only access to the machine.
  config.vm.network "private_network", ip: "192.168.33.10"

  # SSH connections will enable agent forwarding.
  config.ssh.forward_agent = true

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/jeffturcotte/spool"
end
