package main

import (
	"bufio"
  "crypto/tls"
  "crypto/x509"
	"database/sql"
	"flag"
	"fmt"
  "io/ioutil"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/alyu/configparser"

	"github.com/go-sql-driver/mysql"
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
	sslCa          string
	sslCaPath      string
	sslCert        string
	sslCipher      string
	sslKey         string
}

func querydb(host string, cxn *sqlcxn, wg *sync.WaitGroup) {
	defer wg.Done()

  var db *sql.DB
  var err error

  // SSL Support
  if cxn.sslCa != "" || cxn.sslCaPath != "" || cxn.sslCert != "" || cxn.sslCipher != "" || cxn.sslKey != "" {
    rootCAs := x509.NewCertPool()
    {
      pem, err := ioutil.ReadFile(cxn.sslCa)
      if err != nil {
        fmt.Println(err.Error())
      }
      if ok := rootCAs.AppendCertsFromPEM(pem); !ok {
        fmt.Println("Failed to append PEM.")
      }
    }
    clientCerts := make([]tls.Certificate, 0, 1)
    {
      certs, err := tls.LoadX509KeyPair(cxn.sslCert, cxn.sslKey)
      if err != nil {
        fmt.Println(err.Error())
      }
      clientCerts = append(clientCerts, certs)
    }
    mysql.RegisterTLSConfig("omnisql", &tls.Config{
      RootCAs:      rootCAs,
      Certificates: clientCerts,
    })
	  db, err = sql.Open("mysql", cxn.username+":"+cxn.password+"@tcp("+host+":"+strconv.Itoa(cxn.port)+")/?multiStatements=true&tls=omnisql")
  } else {
	  db, err = sql.Open("mysql", cxn.username+":"+cxn.password+"@tcp("+host+":"+strconv.Itoa(cxn.port)+")/?multiStatements=true")
	  if err != nil {
	  	fmt.Println(host, "\t", err.Error())
	  	return
	  }
  }
	defer db.Close()

	// Execute the query
	rows, err := db.Query(cxn.query)
	if err != nil {
		fmt.Println(host, "\t", err.Error())
		return
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println(host, "\t", err.Error())
		return
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	for rows.Next() {
		fmt.Print(host)
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			fmt.Println(host, "\t", err.Error())
			return
		}

		// Now do something with the data.
		// Here we just print each column as a string.
		var value string
		for _, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			fmt.Print("\t", value)
		}
		fmt.Println()
	}
	if err = rows.Err(); err != nil {
		fmt.Println(host, "\t", err.Error())
		return
	}
}

func main() {
	VERSION := "0.0.1"
	var versionBool bool
	var wg sync.WaitGroup
	var databases string
	var defaultsFile string
	hosts := []string{}
	cxn := sqlcxn{}

	flag.StringVar(&cxn.query, "query", "SELECT NOW()", "Query to run (default: 'SELECT NOW()')")
	flag.IntVar(&cxn.port, "port", 3306, "TCP/IP port to connect to (default: 3306)")
	flag.BoolVar(&cxn.all, "all", false, "Run on all databases except i_s, mysql and test (default: false)")
	flag.IntVar(&cxn.threads, "threads", 0, "Number of parallel threads (default: Available CPU Cores)")
	flag.StringVar(&cxn.username, "username", "", "The MySQL user to connect as (default: Current User)")
	flag.StringVar(&defaultsFile, "defaults-file", ".my.cnf", "File to use instead of .my.cnf")
	flag.StringVar(&databases, "databases", "", "Databases (comma-separated) to run query against")
	flag.BoolVar(&versionBool, "version", false, "Display version information and exit")

	flag.Parse()

	if versionBool == true {
		fmt.Println("omnisql VERSION:", VERSION)
		os.Exit(1)
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("Could not determine hosts from STDIN")
		os.Exit(1)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
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

	if _, err := os.Stat(defaultsFile); err == nil {
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
		cxn.socket = section.ValueOf("socket")
		cxn.port, err = strconv.Atoi(section.ValueOf("port"))

		if err != nil {
			cxn.port = 3306
		}

		cxn.sslCa = section.ValueOf("ssl_ca")
		cxn.sslCaPath = section.ValueOf("ssl_capath")
		cxn.sslCert = section.ValueOf("ssl_cert")
		cxn.sslCipher = section.ValueOf("ssl_cipher")
		cxn.sslKey = section.ValueOf("ssl_key")
	}

	// Make sure the SSL stuff is passed as part of the sqlcxn
	for _, host := range hosts {
		wg.Add(1)
		go querydb(host, &cxn, &wg)
	}

	wg.Wait()
}
