package omnisql

import (
	"crypto/tls"
	"database/sql"
	"fmt"
  "log"
	"strconv"
	"sync"

	"github.com/go-sql-driver/mysql"
)

type Sqlcxn struct {
	Query           string
	Port            int
	All             bool
	Escape          bool
	Compress        bool
	ConnectTimeout  int
	ReadTimeout     int
	WriteTimeout    int
	Threads         int
	Username        string
	Password        string
	Socket          string
	Databases       []string
	SslCa           string
	SslCaPath       string
	SslCert         string
	SslCipher       string
	SslKey          string
	MultiStatements bool
	TlsConfig       tls.Config
}

func Query(host string, cxn Sqlcxn, wg *sync.WaitGroup) {
	defer wg.Done()

	var dsn string
	dsn = cxn.Username + ":" + cxn.Password + "@tcp(" + host + ":" + strconv.Itoa(cxn.Port) + ")/"
	if cxn.MultiStatements == true {
		dsn += "?multiStatements=true"
	} else {
		dsn += "?multiStatements=false"
	}
	if cxn.TlsConfig.Certificates != nil {
		mysql.RegisterTLSConfig("omnisql", &cxn.TlsConfig)
		dsn += "&tls=omnisql"
	}
	if cxn.ConnectTimeout <= 0 {
		dsn += "&timeout=2"
	} else {
		dsn += "&timeout=" + strconv.Itoa(cxn.ConnectTimeout) + "s"
	}
	dsn += "&readTimeout=" + strconv.Itoa(cxn.ReadTimeout) + "s"
	dsn += "&writeTimeout=" + strconv.Itoa(cxn.WriteTimeout) + "s"

	// SSL Support
	if cxn.SslCa != "" || cxn.SslCaPath != "" || cxn.SslCert != "" || cxn.SslCipher != "" || cxn.SslKey != "" {
		dsn += "&tls=omnisql"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
    log.Println(host, "\t", err.Error())
		return
	}
	defer db.Close()

	// Execute the query
	rows, err := db.Query(cxn.Query)
	if err != nil {
    log.Println("WARNING: Could not connect to '"+host+"': Unknown MySQL server host '"+host+"'")
		return
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		log.Println(host, "\t", err.Error())
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
		var res string
		res += host
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Println(host, "\t", err.Error())
			return
		}

		// Now do something with the data.
		// Here we just print each column as a string.
		for _, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				res += "\tNULL"
			} else {
				res += "\t"
				res += string(col)
			}
		}
		fmt.Println(res)
	}
	if err = rows.Err(); err != nil {
		log.Println(host, "\t", err.Error())
		return
	}
	return
}
