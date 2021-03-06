package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/alyu/configparser"
	"github.com/rlowe/omnisql"
)

func main() {
	VERSION := "0.0.3"
	var versionBool bool
	var wg sync.WaitGroup
	var databases string
	var defaultsFile string
	hosts := []string{}
	cxn := omnisql.Sqlcxn{}

	flag.StringVar(&cxn.Query, "query", "SELECT NOW()", "Query to run")
	flag.IntVar(&cxn.Port, "port", 3306, "TCP/IP port to connect to (default: 3306)")
	flag.BoolVar(&cxn.All, "all", false, "Run on all databases except i_s, mysql and test (default: false)")
	flag.IntVar(&cxn.Threads, "threads", 0, "Number of parallel threads (default: Available CPU Cores)")
	flag.StringVar(&cxn.Username, "username", "", "The MySQL user to connect as (default: Current User)")
	flag.StringVar(&defaultsFile, "defaults-file", ".my.cnf", "File to use instead of .my.cnf")
	flag.StringVar(&databases, "databases", "", "Databases (comma-separated) to run query against")
	flag.BoolVar(&versionBool, "version", false, "Display version information and exit")
	flag.BoolVar(&cxn.MultiStatements, "multistatements", true, "Allow multiple statements in one query. While this allows batch queries, it also greatly increases the risk of SQL injections. Only the result of the first query is returned, all other results are silently discarded. (default: true)")
	flag.IntVar(&cxn.ConnectTimeout, "connect-timeout", 2, "Connect timeout in seconds")
	flag.IntVar(&cxn.ReadTimeout, "read-timeout", 0, "Read timeout in seconds")
	flag.IntVar(&cxn.WriteTimeout, "write-timeout", 0, "Write timeout in seconds")

	flag.Parse()

	if versionBool == true {
		log.Fatalln("omnisql", VERSION)
	}

	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalln("Could not determine hosts from STDIN")
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		log.Fatalln("Could not determine hosts from STDIN")
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if line != "" {
			hosts = append(hosts, line)
		}
	}
	if len(hosts) < 1 {
		log.Fatalln("Could not determine hosts from STDIN")
	}

	if cxn.Threads == 0 {
		if runtime.GOMAXPROCS(0) < runtime.NumCPU() {
			cxn.Threads = runtime.GOMAXPROCS(0)
		} else {
			cxn.Threads = runtime.NumCPU()
		}
	}

	cxn.Databases = strings.Split(databases, ",")

	u, err := user.Current()
	if err != nil && u == nil {
		log.Println("WARNING: Could not determine current username with user.Current()")
	}
	if cxn.Username == "" {
		cxn.Username = u.Username
	}
	if defaultsFile == ".my.cnf" {
		defaultsFile = u.HomeDir + "/.my.cnf"
	}

	if _, err := os.Stat(defaultsFile); err == nil {
		cxn.ParseDefaultsFile(defaultsFile)
	}

	// SSL Support
	if cxn.SslCa != "" || cxn.SslCaPath != "" || cxn.SslCert != "" || cxn.SslCipher != "" || cxn.SslKey != "" {
		rootCAs := x509.NewCertPool()
		pemPath := filepath.Join(cxn.SslCaPath, cxn.SslCa)
		pem, err := ioutil.ReadFile(pemPath)
		if err != nil {
			log.Fatalln("ERROR: Could Not Read PEM at" + pemPath)
		}
		if ok := rootCAs.AppendCertsFromPEM(pem); !ok {
			log.Fatalln("ERROR: Failed to append PEM to x509.NewCertPool()")
		}
		clientCerts := make([]tls.Certificate, 0, 1)
		certs, err := tls.LoadX509KeyPair(cxn.SslCert, cxn.SslKey)
		if err != nil {
			log.Fatalln("ERROR: Could not load x509 key pair\n" + err.Error())
		}
		clientCerts = append(clientCerts, certs)
		cxn.TlsConfig = tls.Config{
			RootCAs:      rootCAs,
			Certificates: clientCerts,
		}
	}

	for _, host := range hosts {
		wg.Add(1)
		go omnisql.Query(host, cxn, &wg)
	}

	wg.Wait()
}
