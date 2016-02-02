package main

// TODO(0): pmysql SELECT @@version formatting compatibility
// TODO(1): How does pmysql handle multi-statement returns?
// TODO(4): --mysql || --postgres
// TODO(5): Use environmental variables as appropriate
// TODO(5): SSL in go-sql-driver
// TODO(7): --mysql || --postgres

import (
  "bufio"
  "flag"
  "fmt"
  "os"
  "os/user"
  "runtime"
  "strconv"
  "strings"
  "sync"

  "github.com/alyu/configparser"

  _ "github.com/go-sql-driver/mysql"
)

type sqlcxn struct {
  query          string
  port           int
  all            bool
  escape         bool
  compress       bool
  connectTimeout int
  readTimeout    int
  threads        int
  username       string
  password       string
  socket         string
  databases      []string
  ssl_ca         string
  ssl_capath     string
  ssl_cert       string
  ssl_cipher     string
  ssl_key        string
}

// BUG(rlowe): This does nothing!
func querydb(host string, cxn *sqlcxn, wg *sync.WaitGroup) {
  defer wg.Done()
}

func main() {
  var wg sync.WaitGroup
  var databases string
  var defaultsFile string
  hosts := []string{}
  cxn   := sqlcxn{}

  flag.StringVar(&cxn.query, "query", "SELECT NOW()", "Query to run (default: 'SELECT NOW()')")
  flag.IntVar(&cxn.port, "port", 3306, "TCP/IP port to connect to (default: 3306)")
  flag.BoolVar(&cxn.all, "all", false, "Run on all databases except i_s, mysql and test (default: false)")
  flag.IntVar(&cxn.threads, "threads", 0, "Number of parallel threads (default: Available CPU Cores)")
  flag.StringVar(&cxn.username, "username", "", "The MySQL user to connect as (default: Current User)")
  flag.StringVar(&defaultsFile, "defaults-file", ".my.cnf", "File to use instead of .my.cnf")
  flag.StringVar(&databases, "databases", "", "Databases (comma-separated) to run query against")

  flag.Parse()

  fi, err := os.Stdin.Stat()
  if err != nil {
    fmt.Println("Could not determine hosts from STDIN")
    os.Exit(1)
  }
  if fi.Mode() & os.ModeNamedPipe == 0 {
    fmt.Println("Could not determine hosts from STDIN")
    os.Exit(1)
  }

  scanner := bufio.NewScanner(os.Stdin)
  for scanner.Scan() {
    line := strings.Trim(scanner.Text(), " ")
    if line != "" {
      hosts = append(hosts, line)
    }
  }
  if len(hosts) < 1 {
    fmt.Println("Could not determine hosts from STDIN")
    os.Exit(1)
  }

  if cxn.threads == 0 {
    if runtime.GOMAXPROCS(0) < runtime.NumCPU() {
      cxn.threads = runtime.GOMAXPROCS(0)
    } else {
      cxn.threads = runtime.NumCPU()
    }
  }

  cxn.databases = strings.Split(databases, ",")

  u, err := user.Current()
  if err != nil && u == nil {
    fmt.Println(err.Error())
    return
  }
  if cxn.username == "" {
    cxn.username = u.Username
  }
  if defaultsFile == ".my.cnf" {
    defaultsFile = u.HomeDir + "/.my.cnf"
  }

  // Do the defaults file awesomeness!
  cnf, err := configparser.Read(defaultsFile)
  if err != nil {
    // We don't particularly care here
  }

  section, err := cnf.Section("client")
  if err != nil {
    // We don't particularly care here
  }

  if cxn.username == "" {
    cxn.username = section.ValueOf("user")
  }
  if cxn.password == "" {
    cxn.password = section.ValueOf("password")
  }
  cxn.socket   = section.ValueOf("socket")
  cxn.port, err = strconv.Atoi(section.ValueOf("port"))

  if err != nil {
    cxn.port = 3306
  }

  cxn.ssl_ca = section.ValueOf("ssl_ca")
  cxn.ssl_capath = section.ValueOf("ssl_capath")
  cxn.ssl_cert = section.ValueOf("ssl_cert")
  cxn.ssl_cipher = section.ValueOf("ssl_cipher")
  cxn.ssl_key = section.ValueOf("ssl_key")

  // Make sure the SSL stuff is passed as part of the sqlcxn
  for _, host := range hosts {
    wg.Add(1)
    go querydb(host, &cxn, &wg)
  }

  wg.Wait()
}

