/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		wantConfig  *Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config with all fields",
			configData: `{
				"cluster": "imagineer-prod",
				"database": "imagineer",
				"nodes": [
					{
						"name": "primary",
						"hostname": "localhost",
						"port": 5432
					}
				],
				"users": [
					{
						"username": "imagineer_app",
						"password": "secret123",
						"superuser": false
					}
				]
			}`,
			wantConfig: &Config{
				Cluster:  "imagineer-prod",
				Database: "imagineer",
				Nodes: []NodeConfig{
					{
						Name:     "primary",
						Hostname: "localhost",
						Port:     5432,
					},
				},
				Users: []UserConfig{
					{
						Username:  "imagineer_app",
						Password:  "secret123",
						Superuser: false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with multiple nodes",
			configData: `{
				"cluster": "imagineer-cluster",
				"database": "imagineer",
				"nodes": [
					{
						"name": "primary",
						"hostname": "db-primary.local",
						"port": 5432
					},
					{
						"name": "replica",
						"hostname": "db-replica.local",
						"port": 5432
					}
				],
				"users": [
					{
						"username": "admin",
						"password": "adminpass",
						"superuser": true
					},
					{
						"username": "app",
						"password": "apppass",
						"superuser": false
					}
				]
			}`,
			wantConfig: &Config{
				Cluster:  "imagineer-cluster",
				Database: "imagineer",
				Nodes: []NodeConfig{
					{
						Name:     "primary",
						Hostname: "db-primary.local",
						Port:     5432,
					},
					{
						Name:     "replica",
						Hostname: "db-replica.local",
						Port:     5432,
					},
				},
				Users: []UserConfig{
					{
						Username:  "admin",
						Password:  "adminpass",
						Superuser: true,
					},
					{
						Username:  "app",
						Password:  "apppass",
						Superuser: false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal valid config",
			configData: `{
				"cluster": "test",
				"database": "testdb",
				"nodes": [],
				"users": []
			}`,
			wantConfig: &Config{
				Cluster:  "test",
				Database: "testdb",
				Nodes:    []NodeConfig{},
				Users:    []UserConfig{},
			},
			wantErr: false,
		},
		{
			name:        "invalid JSON",
			configData:  `{"cluster": "test", "database":}`,
			wantErr:     true,
			errContains: "failed to parse config file",
		},
		{
			name:       "empty JSON object",
			configData: `{}`,
			wantConfig: &Config{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir, err := os.MkdirTemp("", "imagineer-db-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			// Write config file
			configPath := filepath.Join(tempDir, "config.json")
			err = os.WriteFile(configPath, []byte(tt.configData), 0644)
			require.NoError(t, err)

			// Test LoadConfig
			config, err := LoadConfig(configPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)

			// Verify config matches expected
			assert.Equal(t, tt.wantConfig.Cluster, config.Cluster)
			assert.Equal(t, tt.wantConfig.Database, config.Database)
			assert.Equal(t, len(tt.wantConfig.Nodes), len(config.Nodes))
			assert.Equal(t, len(tt.wantConfig.Users), len(config.Users))

			// Verify nodes
			for i, expectedNode := range tt.wantConfig.Nodes {
				assert.Equal(t, expectedNode.Name, config.Nodes[i].Name)
				assert.Equal(t, expectedNode.Hostname, config.Nodes[i].Hostname)
				assert.Equal(t, expectedNode.Port, config.Nodes[i].Port)
			}

			// Verify users
			for i, expectedUser := range tt.wantConfig.Users {
				assert.Equal(t, expectedUser.Username, config.Users[i].Username)
				assert.Equal(t, expectedUser.Password, config.Users[i].Password)
				assert.Equal(t, expectedUser.Superuser, config.Users[i].Superuser)
			}
		})
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	config, err := LoadConfig("/nonexistent/path/config.json")

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "imagineer-db-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Write empty file
	configPath := filepath.Join(tempDir, "empty.json")
	err = os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	// Test LoadConfig
	config, err := LoadConfig(configPath)

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestLoadConfig_PermissionDenied(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "imagineer-db-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Write config file
	configPath := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, []byte(`{"cluster": "test"}`), 0644)
	require.NoError(t, err)

	// Remove read permissions
	err = os.Chmod(configPath, 0000)
	require.NoError(t, err)
	defer func() { _ = os.Chmod(configPath, 0644) }() // Restore permissions for cleanup

	// Test LoadConfig
	config, err := LoadConfig(configPath)

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestConfig_JSONStructure(t *testing.T) {
	// Test that Config struct can be properly marshaled/unmarshaled
	originalConfig := Config{
		Cluster:  "test-cluster",
		Database: "testdb",
		Nodes: []NodeConfig{
			{
				Name:     "node1",
				Hostname: "localhost",
				Port:     5432,
			},
		},
		Users: []UserConfig{
			{
				Username:  "testuser",
				Password:  "testpass",
				Superuser: true,
			},
		},
	}

	// Marshal
	data, err := json.Marshal(originalConfig)
	require.NoError(t, err)

	// Unmarshal
	var resultConfig Config
	err = json.Unmarshal(data, &resultConfig)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, originalConfig.Cluster, resultConfig.Cluster)
	assert.Equal(t, originalConfig.Database, resultConfig.Database)
	assert.Equal(t, originalConfig.Nodes[0].Name, resultConfig.Nodes[0].Name)
	assert.Equal(t, originalConfig.Nodes[0].Hostname, resultConfig.Nodes[0].Hostname)
	assert.Equal(t, originalConfig.Nodes[0].Port, resultConfig.Nodes[0].Port)
	assert.Equal(t, originalConfig.Users[0].Username, resultConfig.Users[0].Username)
	assert.Equal(t, originalConfig.Users[0].Password, resultConfig.Users[0].Password)
	assert.Equal(t, originalConfig.Users[0].Superuser, resultConfig.Users[0].Superuser)
}

func TestNodeConfig_JSONTags(t *testing.T) {
	node := NodeConfig{
		Name:     "primary",
		Hostname: "db.example.com",
		Port:     5433,
	}

	data, err := json.Marshal(node)
	require.NoError(t, err)

	// Verify JSON field names
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Equal(t, "primary", jsonMap["name"])
	assert.Equal(t, "db.example.com", jsonMap["hostname"])
	assert.Equal(t, float64(5433), jsonMap["port"])
}

func TestUserConfig_JSONTags(t *testing.T) {
	user := UserConfig{
		Username:  "admin",
		Password:  "secretpass",
		Superuser: true,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	// Verify JSON field names
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Equal(t, "admin", jsonMap["username"])
	assert.Equal(t, "secretpass", jsonMap["password"])
	assert.Equal(t, true, jsonMap["superuser"])
}
