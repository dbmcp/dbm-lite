package main

import (
	"fmt"

	"dbm-lite/internal/dbtype"
)

func main() {
	fmt.Println("=== TiDB Cloud SSL Test ===")

	// Test 1: TiDB Cloud with SSL enabled, NO CA file (should use system CA via tls=true)
	fmt.Println("\n[Test 1] TiDB Cloud + SSL ON, NO CA file (expect tls=true)")
	p1 := &dbtype.ConnectionParams{
		Type:     "tidb",
		Host:     "gateway01.us-west-2.prod.aws.tidbcloud.com",
		Port:     4000,
		Username: "user",
		Password: "password",
		Database: "test",
		SSLMode:  "true",
	}
	dsn1, _, err1 := dbtype.BuildDSN(p1)
	if err1 != nil {
		fmt.Printf("  FAIL - Error: %v\n", err1)
	} else {
		fmt.Printf("  OK   - DSN: %s\n", dsn1)
	}

	// Test 2: TiDB Cloud with SSL + CA file
	fmt.Println("\n[Test 2] TiDB Cloud + SSL ON + CA file path (expect tls=dbmlite_xxx)")
	p2 := &dbtype.ConnectionParams{
		Type:      "tidb",
		Host:      "gateway01.us-west-2.prod.aws.tidbcloud.com",
		Port:      4000,
		Username:  "user",
		Password:  "password",
		Database:  "test",
		SSLMode:   "true",
		SSLCAFile: "/etc/ssl/certs/ca-certificates.crt",
	}
	dsn2, _, err2 := dbtype.BuildDSN(p2)
	if err2 != nil {
		fmt.Printf("  FAIL - Error: %v\n", err2)
	} else {
		fmt.Printf("  OK   - DSN: %s\n", dsn2)
	}

	// Test 3: MySQL without SSL
	fmt.Println("\n[Test 3] MySQL without SSL (expect NO tls param)")
	p3 := &dbtype.ConnectionParams{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "root",
		Database: "test",
		SSLMode:  "false",
	}
	dsn3, _, err3 := dbtype.BuildDSN(p3)
	if err3 != nil {
		fmt.Printf("  FAIL - Error: %v\n", err3)
	} else {
		fmt.Printf("  OK   - DSN: %s\n", dsn3)
	}

	// Test 4: SSL mode "require"
	fmt.Println("\n[Test 4] MySQL + SSL mode 'require' (expect tls=true)")
	p4 := &dbtype.ConnectionParams{
		Type:     "mysql",
		Host:     "mysql.example.com",
		Port:     3306,
		Username: "user",
		Password: "pass",
		Database: "test",
		SSLMode:  "require",
	}
	dsn4, _, err4 := dbtype.BuildDSN(p4)
	if err4 != nil {
		fmt.Printf("  FAIL - Error: %v\n", err4)
	} else {
		fmt.Printf("  OK   - DSN: %s\n", dsn4)
	}

	// Test 5: SSL ON with PEM content as CA
	fmt.Println("\n[Test 5] MySQL + SSL ON + CA PEM content (expect tls=dbmlite_xxx)")
	p5 := &dbtype.ConnectionParams{
		Type:      "mysql",
		Host:      "mysql.example.com",
		Port:      3306,
		Username:  "user",
		Password:  "pass",
		Database:  "test",
		SSLMode:   "true",
		SSLCAFile: "-----BEGIN CERTIFICATE-----\nMIIDazCCAlOgAwIBAgIUXXXXX\n-----END CERTIFICATE-----",
	}
	dsn5, _, err5 := dbtype.BuildDSN(p5)
	if err5 != nil {
		fmt.Printf("  FAIL - Error: %v\n", err5)
	} else {
		fmt.Printf("  OK   - DSN: %s\n", dsn5)
	}

	fmt.Println("\n=== All tests completed ===")
}
