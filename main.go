package main

import (
  "flag"
  "fmt"
  "os/user"
  "runtime"
)

func main() {
  var query string
  var port int
  var all bool
  var escape bool
  var compress bool
  var connectTimeout int
  var readTimeout int
  var threads int
  var username string
  var password string
  var socket string
  var defaultsFile string

  databases := []string{}
  hosts     := []string{}

  // Command-Line options to override the defaults
  flag.StringVar(&query, "query", "SELECT NOW()", "Query to run (default: 'SELECT NOW()')")
  flag.IntVar(&port, "port", 3306, "TCP/IP port to connect to (default: 3306)")
  flag.BoolVar(&all, "all", false, "Run on all databases except i_s, mysql and test (default: false)")
  flag.IntVar(&threads, "threads", 0, "Number of parallel threads (default: Available CPU Cores)")
  flag.StringVar(&username, "username", "", "The MySQL user to connect as (default: Current User)")

  flag.Parse()

  if username == "" {
    u, err := user.Current()
    if err != nil && u == nil {
      fmt.Println(err.Error())
      return
    }
    username = u.Username
  }

  if threads == 0 {
    if runtime.GOMAXPROCS(0) < runtime.NumCPU() {
      threads = runtime.GOMAXPROCS(0)
    } else {
      threads = runtime.NumCPU()
    }
  }

  for _, host := range hosts {
    fmt.Println(host)
  }

  for _, database := range databases {
    fmt.Println(database)
  }

  fmt.Println("Query:", query)
  fmt.Println("Port:", port)
  fmt.Println("All:", all)
  fmt.Println("Threads:", threads)
  fmt.Println("Escape:", escape)
  fmt.Println("Compress:", compress)
  fmt.Println("Connect Timeout:", connectTimeout)
  fmt.Println("Read Timeout:", readTimeout)
  fmt.Println("Username:", username)
  fmt.Println("Password:", password)
  fmt.Println("Defaults File:", defaultsFile)
  fmt.Println("Socket:", socket)
}

