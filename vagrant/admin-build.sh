echo "GOPATH=/omnisql" >> ~vagrant/.bash_profile
echo "export GOPATH" >> ~vagrant/.bash_profile

# Basic Development
sudo yum -y install vim
curl -L https://bit.ly/janus-bootstrap | bash

export GOPATH="/omnisql"
cd $GOPATH
go get github.com/alyu/configparser
go get github.com/go-sql-driver/mysql
