#!/bin/bash

if [[ -e /etc/redhat-release ]]; then
  # Percona's Yum Repository
  sudo yum -y install http://www.percona.com/downloads/percona-release/redhat/0.1-3/percona-release-0.1-3.noarch.rpm

  # All the project dependencies to build
  sudo yum -y install ruby-devel gcc rpm-build git
  gem install fpm

  # Go (via EPEL)
  sudo yum -y install epel-release
  sudo yum -y install golang
elif [[ -e /etc/debian_version ]]; then
  sudo echo exit 101 > /usr/sbin/policy-rc.d
  sudo chmod +x /usr/sbin/policy-rc.d
  # Percona's Apt Repository
  sudo apt-key adv --keyserver keys.gnupg.net --recv-keys 1C4CBDCDCD2EFD2A
  echo "deb http://repo.percona.com/apt "$(lsb_release -sc)" main" | sudo tee /etc/apt/sources.list.d/percona.list
  sudo apt-get -y update
  sudo apt-get -y install debconf-utils
  echo "golang-go golang-go/dashboard boolean true" | sudo debconf-set-selections
  #echo "percona-server-server-5.6 percona-server-server/root_password password vagrant" | sudo debconf-set-selections
  #echo "percona-server-server-5.6 percona-server-server/root_password_again password vagrant" | sudo debconf-set-selections
  DEBIAN_FRONTEND=noninteractive

  # All the project dependencies to build
  sudo apt-get -y install ruby-dev gcc git rubygems rpm
  sudo gem install fpm

  # Go
  sudo apt-get -y install golang-go
fi

cat <<-EOF >> /etc/hosts
  192.168.57.200   admin
  192.168.57.201   db1
  192.168.57.202   db2
  192.168.57.203   db3
  192.168.57.204   db4
EOF

# Generated a random SSH keypair to be used by the vagrant user for convenience
mkdir -p /home/vagrant/.ssh

cp /omnisql/vagrant/vagrant-ssh-key /home/vagrant/.ssh/id_rsa
cp /omnisql/vagrant/vagrant-ssh-key.pub /home/vagrant/.ssh/id_rsa.pub

cat <<EOF > /home/vagrant/.ssh/config
Host admin
  User vagrant
  IdentifyFile /home/vagrant/.ssh/id_rsa
Host db1
  User vagrant
  IdentifyFile /home/vagrant/.ssh/id_rsa
Host db2
  User vagrant
  IdentifyFile /home/vagrant/.ssh/id_rsa
Host db3
  User vagrant
  IdentifyFile /home/vagrant/.ssh/id_rsa
Host db4
  User vagrant
  IdentifyFile /home/vagrant/.ssh/id_rsa
EOF

chmod go-rwx /home/vagrant/.ssh/*
chown -R vagrant:vagrant /home/vagrant/.ssh

if [[ -e /etc/redhat-release ]]; then
  sudo service iptables stop
fi

if [[ $HOSTNAME == 'admin' ]]; then
  bash /omnisql/vagrant/admin-build.sh
  if [[ -e /omnisql/vagrant/admin-post-install.sh ]]; then
    bash /omnisql/vagrant/admin-post-install.sh
  fi
else
  bash /omnisql/vagrant/$HOSTNAME-build.sh

  if [[ -e /omnisql/vagrant/db-post-install.sh ]]; then
    bash /omnisql/vagrant/db-post-install.sh
  fi

  if [[ -e /omnisql/vagrant/$HOSTNAME-post-install.sh ]]; then
    bash /omnisql/vagrant/$HOSTNAME-post-install.sh
  fi
fi
