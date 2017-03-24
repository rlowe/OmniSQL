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

// Sqlcxn contains all information to successfully connect to MySQL
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
	TlsConfig       *tls.Config
}

// Dsn builds a go-sql-driver/mysql DSN for connecting to MySQL
func (s *Sqlcxn) Dsn() string {
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

	return dsn
}

// ParseDefaultsFile updates a Sqlcxn with any parameters in the defaults file
func (s *Sqlcxn) ParseDefaultsFile(string defaultsFile) {
	// Do the defaults file awesomeness!
	cnf, err := configparser.Read(defaultsFile)
	if err != nil {
		// We don't particularly care here
	}

	section, _ := cnf.Section("client")

	if cxn.Username == "" {
		cxn.Username = section.ValueOf("user")
	}
	if cxn.Password == "" {
		cxn.Password = section.ValueOf("password")
	}
	cxn.Socket = section.ValueOf("socket")
	cxn.Port, err = strconv.Atoi(section.ValueOf("port"))

	if err != nil {
		cxn.Port = 3306
	}

	// dash and underscore are equivalent
	cxn.SslCa = section.ValueOf("ssl_ca")
	if cxn.SslCa == "" {
		cxn.SslCa = section.ValueOf("ssl-ca")
	}
	cxn.SslCaPath = section.ValueOf("ssl_capath")
	if cxn.SslCaPath == "" {
		cxn.SslCaPath = section.ValueOf("ssl-capath")
	}
	cxn.SslCert = section.ValueOf("ssl_cert")
	if cxn.SslCert == "" {
		cxn.SslCert = section.ValueOf("ssl-cert")
	}
	cxn.SslCipher = section.ValueOf("ssl_cipher")
	if cxn.SslCipher == "" {
		cxn.SslCipher = section.ValueOf("ssl-cipher")
	}
	cxn.SslKey = section.ValueOf("ssl_key")
	if cxn.SslKey == "" {
		cxn.SslKey = section.ValueOf("ssl-key")
	}
}

// Query does all the work, sending queries to the database and coordinating
func Query(host string, cxn Sqlcxn, wg *sync.WaitGroup) {
	defer wg.Done()

	dsn := cxn.Dsn()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Println(host, "\t", err.Error())
		return
	}
	defer db.Close()

	// Execute the query
	rows, err := db.Query(cxn.Query)
	if err != nil {
		log.Println("WARNING: Could not connect to '" + host + "': Unknown MySQL server host '" + host + "'")
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
