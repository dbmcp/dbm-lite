package dbtype

import (
	"fmt"
	"strings"
	"testing"
)

func TestDSNSSLBuildScenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		params         *ConnectionParams
		wantContain    string
		wantNotContain []string
	}{
		{
			name: "[Test 1] TiDB Cloud + SSL ON, NO CA file (expect tls=true)",
			params: &ConnectionParams{
				Type:     "tidb",
				Host:     "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port:     4000,
				Username: "user",
				Password: "password",
				Database: "test",
				SSLMode:  "true",
			},
			wantContain:    "tls=true",
			wantNotContain: []string{"tls=skip-verify", "tls=dbmlite_"},
		},
		{
			name: "[Test 2] TiDB Cloud + SSL ON + CA file path (expect tls=dbmlite_xxx)",
			params: &ConnectionParams{
				Type:      "tidb",
				Host:      "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port:      4000,
				Username:  "user",
				Password:  "password",
				Database:  "test",
				SSLMode:   "true",
				SSLCAFile: "/etc/ssl/certs/ca-certificates.crt",
			},
			wantContain:    "tls=dbmlite_",
			wantNotContain: []string{"tls=true", "tls=skip-verify"},
		},
		{
			name: "[Test 3] MySQL without SSL (expect NO tls param)",
			params: &ConnectionParams{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "root",
				Database: "test",
				SSLMode:  "false",
			},
			wantContain:    "",
			wantNotContain: []string{"tls="},
		},
		{
			name: "[Test 4] MySQL + SSL mode 'require' (expect tls=true)",
			params: &ConnectionParams{
				Type:     "mysql",
				Host:     "mysql.example.com",
				Port:     3306,
				Username: "user",
				Password: "pass",
				Database: "test",
				SSLMode:  "require",
			},
			wantContain:    "tls=true",
			wantNotContain: []string{"tls=skip-verify", "tls=dbmlite_"},
		},
		{
			name: "[Test 5] MySQL + SSL ON + CA PEM content (expect tls=dbmlite_xxx)",
			params: &ConnectionParams{
				Type:      "mysql",
				Host:      "mysql.example.com",
				Port:      3306,
				Username:  "user",
				Password:  "pass",
				Database:  "test",
				SSLMode:   "true",
				SSLCAFile: "-----BEGIN CERTIFICATE-----\nMIIDazCCAlOgAwIBAgIUXXXXX\n-----END CERTIFICATE-----",
			},
			wantContain:    "tls=dbmlite_",
			wantNotContain: []string{"tls=true", "tls=skip-verify"},
		},
	}

	fmt.Println("=== TiDB Cloud SSL Test ===")
	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			dsn, _, err := buildDSN(sc.params)
			if err != nil {
				t.Fatalf("FAIL - buildDSN returned error: %v", err)
			}

			// Positive check
			if sc.wantContain != "" && !strings.Contains(dsn, sc.wantContain) {
				t.Errorf("FAIL - Expected DSN to contain %q, but DSN was: %s", sc.wantContain, dsn)
			}

			// Negative check
			for _, nc := range sc.wantNotContain {
				if strings.Contains(dsn, nc) {
					t.Errorf("FAIL - Expected DSN NOT to contain %q, but DSN was: %s", nc, dsn)
				}
			}

			fmt.Printf("\n%s\n  OK   - DSN: %s\n", sc.name, dsn)
		})
	}
	fmt.Println("\n=== All Tests Completed ===")
}
