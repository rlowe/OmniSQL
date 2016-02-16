if [[ -e /etc/redhat-release ]]; then
  sudo yum -y install Percona-Server-client-57 Percona-Server-devel-57 Percona-Server-server-57 Percona-Server-shared-57
  sudo rm -rf /etc/my.cnf
  sudo cp /omnisql/vagrant/db4-my.cnf /etc/my.cnf
  sudo mysqld --initialize-insecure --explicit_defaults_for_timestamp
  sudo service mysql start
elif [[ -e /etc/debian_version ]]; then
  sudo cp /omnisql/vagrant/db4-my.cnf /etc/mysql/my.cnf
  sudo /etc/init.d/mysql restart

sudo cat <<-EOF >> /root/.my.cnf
  [client]
  user     = root
  password = vagrant
EOF
fi

/usr/bin/mysql -uroot -ss -e 'GRANT REPLICATION SLAVE ON *.* TO "repl"@"192.168.57.%" IDENTIFIED BY "vagrant_repl"'
/usr/bin/mysql -uroot -ss -e 'CHANGE MASTER TO MASTER_HOST="192.168.57.202", MASTER_USER="repl", MASTER_PASSWORD="vagrant_repl"'
/usr/bin/mysql -uroot -ss -e 'CREATE USER vagrant@"%"'
/usr/bin/mysql -uroot -ss -e 'GRANT ALL PRIVILEGES ON *.* TO vagrant@"%"'
/usr/bin/mysql -uroot -ss -e 'START SLAVE'
