#!/bin/bash

# Simple packaging of Omnisql
#
# Requires fpm: https://github.com/jordansissel/fpm

epoch="1"
release_version="0.0.2"
release_dir=/tmp/omnisql-release
release_files_dir=$release_dir/omnisql
rm -rf $release_dir/*
mkdir -p $release_files_dir/
mkdir -p $release_files_dir/usr/local/bin

cd  $(dirname $0)
for f in $(find . -name "*.go"); do go fmt $f; done

GOPATH=$GOPATH:$(pwd)
go get github.com/alyu/configparser
go get github.com/go-sql-driver/mysql
go build -o $release_files_dir/usr/local/omnisql main.go

if [[ $? -ne 0 ]] ; then
	exit 1
fi

cd $release_dir
# tar packaging
tar cfz omnisql-"${release_version}".tar.gz ./omnisql
# rpm packaging
fpm --epoch $epoch -v "${release_version}" -f -s dir -t rpm -n omnisql -C $release_dir/omnisql --prefix=/ .
fpm --epoch $epoch --deb-no-default-config-files -v "${release_version}" -f -s dir -t deb -n omnisql -C $release_dir/omnisql --prefix=/ .

echo "---"
echo "Done. Find releases in $release_dir"
