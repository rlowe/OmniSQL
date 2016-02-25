echo "GOPATH=/omnisql" >> ~vagrant/.bash_profile
echo "export GOPATH" >> ~vagrant/.bash_profile

export GOPATH="/omnisql"
cd $GOPATH
go get github.com/alyu/configparser
go get github.com/go-sql-driver/mysql
