/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/golang/glog"

	api "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/util/sysctl"

	"k8s.io/ingress-nginx/internal/ingress"
)

// newUpstream creates an upstream without servers.
func newUpstream(name string) *ingress.Backend {
	return &ingress.Backend{
		Name:      name,
		Endpoints: []ingress.Endpoint{},
		Service:   &api.Service{},
		SessionAffinity: ingress.SessionAffinityConfig{
			CookieSessionAffinity: ingress.CookieSessionAffinity{
				Locations: make(map[string][]string),
			},
		},
	}
}

// sysctlSomaxconn returns the maximum number of connections that can be queued
// for acceptance (value of net.core.somaxconn)
// http://nginx.org/en/docs/http/ngx_http_core_module.html#listen
func sysctlSomaxconn() int {
	maxConns, err := sysctl.New().GetSysctl("net/core/somaxconn")
	if err != nil || maxConns < 512 {
		glog.V(3).Infof("net.core.somaxconn=%v (using system default)", maxConns)
		return 511
	}

	return maxConns
}

// sysctlFSFileMax returns the maximum number of open file descriptors (value
// of fs.file-max) or 0 in case of error.
func sysctlFSFileMax() int {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		glog.Errorf("Error reading system maximum number of open file descriptors (RLIMIT_NOFILE): %v", err)
		return 0
	}
	glog.V(2).Infof("rlimit.max=%v", rLimit.Max)
	return int(rLimit.Max)
}

const (
	defBinary = "/usr/sbin/nginx"
	cfgPath   = "/etc/nginx/nginx.conf"
)

func nginxExecCommand(args ...string) *exec.Cmd {
	ngx := os.Getenv("NGINX_BINARY")
	if ngx == "" {
		ngx = defBinary
	}

	cmdArgs := []string{"-c", cfgPath}
	cmdArgs = append(cmdArgs, args...)
	return exec.Command(ngx, cmdArgs...)
}

func nginxTestCommand(cfg string) *exec.Cmd {
	ngx := os.Getenv("NGINX_BINARY")
	if ngx == "" {
		ngx = defBinary
	}

	return exec.Command(ngx, "-c", cfg, "-t")
}
