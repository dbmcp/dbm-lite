package dbtype

import (
	"strings"
	"testing"
)

func TestBuildDSNWithSSL(t *testing.T) {
	tests := []struct {
		name     string
		params   *ConnectionParams
		wantInDSN []string
	}{
		{
			name: "TiDB with SSL enabled but no CA file should use tls=true (system CA)",
			params: &ConnectionParams{
				Type:     "tidb",
				Host:     "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port:     4000,
				Username: "user",
				Password: "pass",
				Database: "test",
				SSLMode:  "true",
			},
			wantInDSN: []string{"tls=true"},
		},
		{
			name: "MySQL with SSL enabled but no CA file should use tls=true",
			params: &ConnectionParams{
				Type:     "mysql",
				Host:     "mysql.example.com",
				Port:     3306,
				Username: "user",
				Password: "pass",
				Database: "test",
				SSLMode:  "true",
			},
			wantInDSN: []string{"tls=true"},
		},
		{
			name: "TiDB Cloud with SSL and CA file path",
			params: &ConnectionParams{
				Type:      "tidb",
				Host:      "gateway01.us-west-2.prod.aws.tidbcloud.com",
				Port:      4000,
				Username:  "user",
				Password:  "pass",
				Database:  "test",
				SSLMode:   "true",
				SSLCAFile: "/etc/ssl/certs/ca-certificates.crt",
			},
			wantInDSN: []string{"tls=dbmlite_"},
		},
		{
			name: "No SSL should not include tls param",
			params: &ConnectionParams{
				Type:     "mysql",
				Host:     "localhost",
				Port:     3306,
				Username: "root",
				Password: "root",
				Database: "test",
				SSLMode:  "false",
			},
			wantInDSN: []string{},
		},
		{
			name: "SSL require mode should use tls=true",
			params: &ConnectionParams{
				Type:     "mysql",
				Host:     "mysql.example.com",
				Port:     3306,
				Username: "user",
				Password: "pass",
				Database: "test",
				SSLMode:  "require",
			},
			wantInDSN: []string{"tls=true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, _, err := buildDSN(tt.params)
			if err != nil {
				t.Fatalf("buildDSN returned error: %v", err)
			}
			for _, want := range tt.wantInDSN {
				if want == "" {
					// negative check: no tls param
					if strings.Contains(dsn, "tls=") {
						t.Errorf("expected no tls param but DSN was: %s", dsn)
					}
				} else {
					if !strings.Contains(dsn, want) {
						t.Errorf("expected DSN to contain %q but DSN was: %s", want, dsn)
					}
				}
			}
			// For no-ssl cases, ensure no tls= param
			if len(tt.wantInDSN) == 0 && strings.Contains(dsn, "tls=") {
				t.Errorf("expected no tls param but DSN was: %s", dsn)
			}
			t.Logf("DSN: %s", dsn)
		})
	}
}
