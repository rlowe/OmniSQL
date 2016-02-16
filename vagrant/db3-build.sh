if [[ -e /etc/redhat-release ]]; then
  sudo yum -y install Percona-Server-client-56 Percona-Server-devel-56 Percona-Server-server-56 Percona-Server-shared-56
  sudo rm -rf /etc/my.cnf
  sudo cp /omnisql/vagrant/db3-my.cnf /etc/my.cnf
  sudo service mysql start
elif [[ -e /etc/debian_version ]]; then
  sudo cp /omnisql/vagrant/db3-my.cnf /etc/mysql/my.cnf
  sudo /etc/init.d/mysql restart

sudo cat <<-EOF >> /root/.my.cnf
  [client]
  user     = root
  password = vagrant
EOF
fi

/usr/bin/mysql -uroot -ss -e 'GRANT REPLICATION SLAVE ON *.* TO "repl"@"192.168.57.%" IDENTIFIED BY "vagrant_repl"'
/usr/bin/mysql -uroot -ss -e 'CHANGE MASTER TO MASTER_HOST="192.168.57.201", MASTER_USER="repl", MASTER_PASSWORD="vagrant_repl"'
/usr/bin/mysql -uroot -ss -e 'GRANT ALL PRIVILEGES ON *.* TO vagrant@"%"'
/usr/bin/mysql -uroot -ss -e 'START SLAVE'
