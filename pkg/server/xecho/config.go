// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xecho

import (
	"fmt"

	"github.com/grey0904/jengine/pkg/conf"
	"github.com/grey0904/jengine/pkg/core/constant"
	"github.com/grey0904/jengine/pkg/core/ecode"
	"github.com/grey0904/jengine/pkg/flag"
	"github.com/grey0904/jengine/pkg/xlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Config HTTP config
type Config struct {
	Host            string
	Port            int
	Deployment      string
	Debug           bool
	DisableMetric   bool
	DisableTrace    bool
	DisableSentinel bool
	CertFile        string
	PrivateFile     string
	EnableTLS       bool

	SlowQueryThresholdInMilli int64

	logger *xlog.Logger
}

// DefaultConfig ...
func DefaultConfig() *Config {
	return &Config{
		Host:                      flag.String("host"),
		Port:                      9091,
		Debug:                     false,
		Deployment:                constant.DefaultDeployment,
		SlowQueryThresholdInMilli: 500, // 500ms
		logger:                    xlog.Jupiter().Named(ecode.ModEchoServer),
		EnableTLS:                 false,
		CertFile:                  "cert.pem",
		PrivateFile:               "private.pem",
	}
}

// StdConfig Jupiter Standard HTTP Server config
func StdConfig(name string) *Config {
	return RawConfig(constant.ConfigKey("server." + name))
}

// RawConfig ...
func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.UnmarshalKey(key, &config); err != nil &&
		errors.Cause(err) != conf.ErrInvalidKey {
		config.logger.Panic("http server parse config panic", xlog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), xlog.FieldErr(err), xlog.FieldKey(key), xlog.FieldValueAny(config))
	}
	return config
}

// WithLogger ...
func (config *Config) WithLogger(logger *xlog.Logger) *Config {
	config.logger = logger
	return config
}

// WithHost ...
func (config *Config) WithHost(host string) *Config {
	config.Host = host
	return config
}

// WithPort ...
func (config *Config) WithPort(port int) *Config {
	config.Port = port
	return config
}

func (config *Config) MustBuild() *Server {
	server, err := config.Build()
	if err != nil {
		config.logger.Panic("build echo server failed", zap.Error(err))
	}
	return server
}

// Build create server instance, then initialize it with necessary interceptor
func (config *Config) Build() (*Server, error) {
	server, err := newServer(config)
	if err != nil {
		return nil, err
	}
	server.Use(recoverMiddleware(config.SlowQueryThresholdInMilli))

	if !config.DisableMetric {
		server.Use(metricServerInterceptor())
	}

	if !config.DisableTrace {
		server.Use(traceServerInterceptor())
	}

	if !config.DisableSentinel {
		server.Use(sentinelServerInterceptor())
	}

	return server, nil
}

// Address ...
func (config *Config) Address() string {
	return fmt.Sprintf("%s:%d", config.Host, config.Port)
}
