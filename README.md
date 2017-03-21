# OmniSQL - Multi-Server MySQL Command-Line Client

[![Go Report Card](https://goreportcard.com/badge/github.com/rlowe/omnisql)](https://goreportcard.com/report/github.com/rlowe/omnisql)


# Testing in Vagrant

    OmniSQL> vagrant up
    OmniSQL> vagrant ssh admin
    Welcome to your Vagrant-built virtual machine.
    [vagrant@admin ~]$ cd /omnisql/
    # Note that without a --query, SELECT NOW() is used
    [vagrant@admin omnisql]$ cat vagrant/hosts |go run main.go 
    db2 2016-02-15 23:59:24
    db1 2016-02-15 23:59:24
    db4 2016-02-15 23:59:23
    db3 2016-02-15 23:59:23
    [vagrant@admin omnisql]$

# Building RPMs

# License

[Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0)
