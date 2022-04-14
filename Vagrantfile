# -*- mode: ruby -*-
# vi: set ft=ruby :

# NOTES:
# * current (hcdiag) directory is implicitly mounted to /vagrant in the VM
# * for vault:
#   export VAULT_SKIP_VERIFY=1
#   export VAULT_TOKEN=localdev

Vagrant.configure("2") do |config|
  # boxes at https://vagrantcloud.com/search
  config.vm.box = "ubuntu/focal64"

  config.vm.network "forwarded_port", guest: 4646, host: 4646 # nomad
  config.vm.network "forwarded_port", guest: 8200, host: 8200 # vault
  config.vm.network "forwarded_port", guest: 8500, host: 8500 # consul

  config.vm.provision "install misc utilities", type: "shell", inline: <<-SHELL
    pkgs='jq make'
    dpkg -l $pkgs >/dev/null && {
      echo "$pkgs are already installed"
      exit
    }
    apt-get update && apt-get install -y $pkgs
  SHELL

  config.vm.provision "install golang", type: "shell", inline: <<-SHELL
    go version 2>/dev/null && exit
    set -xe
    version=1.17.7
    curl -LsSo go$version.tar.gz https://go.dev/dl/go$version.linux-amd64.tar.gz
    tar -C /usr/local -xzf go$version.tar.gz
    echo 'export PATH="/usr/local/go/bin:$PATH"' > /etc/profile.d/golang.sh
    source /etc/profile.d/golang.sh
    go version
  SHELL

  # e.g. https://learn.hashicorp.com/tutorials/vault/getting-started-install
  config.vm.provision "install our softwares", type: "shell", inline: <<-SHELL
    pkgs='consul nomad vault'
    dpkg -l $pkgs >/dev/null && {
      echo "$pkgs are already installed"
      exit
    }

    # add key and repo
    curl -fsSL https://apt.releases.hashicorp.com/gpg | apt-key add -
    apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"

    # do the install
    apt-get update && apt-get install -y $pkgs

    # consul apt package's systemd unit expects this file, but doesn't provide it..
    sudo -u consul touch /etc/consul.d/consul.env
    # also make consul a server, and set its client addr so it can be reached from the host.
    cat <<CONSUL_CONFIG > /etc/consul.d/consul-vagrant.hcl
server = true
bootstrap_expect = 1
client_addr = "0.0.0.0"
CONSUL_CONFIG

    # enable and start
    for p in $pkgs; do
      systemctl enable $p
      systemctl start $p
    done
  SHELL

  config.vm.provision "initialize vault", type: "shell", inline: <<-SHELL
    # apt package's provided self-signed cert means we need this for vault CLI to work
    export VAULT_SKIP_VERIFY=1

    vault status | grep -q 'Sealed.*false' && {
      echo 'vault is already unsealed'
      exit
    }

    # wait for vault to start
    for _ in {1..5}; do
      vault status | grep -q . && break
      sleep 2
    done

    # initalize
    # storing this in plaintext is bad, but we're in vagrant, so w/e
    test -f /opt/vault/init || vault operator init > /opt/vault/init

    # add script to unseal automatically on init (i.e. reboot)
    cat <<EOF > /usr/local/bin/vault-unseal
#!/bin/bash
export VAULT_SKIP_VERIFY=1
for _ in {1..10}; do
  vault status | grep . >/dev/null && break
  sleep 3
done
awk '/^Unseal/ {print \\$NF}' /opt/vault/init | while read -r k; do
  vault status | grep 'Sealed.*false' && break
  vault operator unseal "\\$k" | grep -i seal
done
EOF
    chmod +x /usr/local/bin/vault-unseal
    # unseal on reboot
    echo '@reboot vault /usr/local/bin/vault-unseal' > /etc/cron.d/vault-unseal
    # unseal now
    vault-unseal

    # convenience vault token = "localdev"
    vault token lookup localdev >/dev/null || {
      awk '/Root Token/ {print$NF}' /opt/vault/init > /root/.vault-token
      vault token create -display-name localdev -id localdev
      echo localdev > /home/vagrant/.vault-token
    }
  SHELL

  config.vm.provision "setup user profile", type: "shell", inline: <<-SHELL
    cat <<EOF > /etc/profile.d/hashicorp.sh
export PATH=/vagrant/bin:$PATH
export VAULT_SKIP_VERIFY=1
EOF
  SHELL

  config.vm.provision "build hcdiag", type: "shell", inline: <<-SHELL
    cd /vagrant
    make clean build
  SHELL
end
