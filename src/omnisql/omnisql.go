package omnisql

import (
  "crypto/tls"
  "crypto/x509"
	"database/sql"
  "fmt"
  "io/ioutil"
  "strconv"
  "sync"

  "github.com/go-sql-driver/mysql"
)

type Sqlcxn struct {
	Query          string
	Port           int
	All            bool
	Escape         bool
	Compress       bool
	ConnectTimeout int
	ReadTimeout    int
	Threads        int
	Username       string
	Password       string
	Socket         string
	Databases      []string
	SslCa          string
	SslCaPath      string
	SslCert        string
	SslCipher      string
	SslKey         string
}

func Query(host string, cxn *Sqlcxn, wg *sync.WaitGroup) {
	defer wg.Done()

  var db *sql.DB
  var err error

  // SSL Support
  if cxn.SslCa != "" || cxn.SslCaPath != "" || cxn.SslCert != "" || cxn.SslCipher != "" || cxn.SslKey != "" {
    rootCAs := x509.NewCertPool()
    {
      pem, err := ioutil.ReadFile(cxn.SslCa)
      if err != nil {
        fmt.Println(err.Error())
      }
      if ok := rootCAs.AppendCertsFromPEM(pem); !ok {
        fmt.Println("Failed to append PEM.")
      }
    }
    clientCerts := make([]tls.Certificate, 0, 1)
    {
      certs, err := tls.LoadX509KeyPair(cxn.SslCert, cxn.SslKey)
      if err != nil {
        fmt.Println(err.Error())
      }
      clientCerts = append(clientCerts, certs)
    }
    mysql.RegisterTLSConfig("omnisql", &tls.Config{
      RootCAs:      rootCAs,
      Certificates: clientCerts,
    })
	  db, err = sql.Open("mysql", cxn.Username+":"+cxn.Password+"@tcp("+host+":"+strconv.Itoa(cxn.Port)+")/?multiStatements=true&tls=omnisql")
  } else {
	  db, err = sql.Open("mysql", cxn.Username+":"+cxn.Password+"@tcp("+host+":"+strconv.Itoa(cxn.Port)+")/?multiStatements=true")
	  if err != nil {
	  	fmt.Println(host, "\t", err.Error())
	  	return
	  }
  }
	defer db.Close()

	// Execute the query
	rows, err := db.Query(cxn.Query)
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
