/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package detect

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/klauspost/cpuid"
	"github.com/spf13/viper"
	"golang.org/x/sys/cpu"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// RuntimeOS returns the runtime operating system
func RuntimeOS() string {
	return runtime.GOOS
}

// RuntimeArch returns the runtime architecture
func RuntimeArch() string {
	arch := runtime.GOARCH
	if arch == "arm" {
		// runtime.GOARM
		if !cpu.ARM.HasVFP {
			return "arm/v5"
		}
		if !cpu.ARM.HasVFPv3 {
			return "arm/v6"
		}
		// "arm" (== "arm/v7")
	}
	return arch
}

// IsMicrosoftWSL will return true if process is running in WSL in windows
// checking for WSL env var based on this https://github.com/microsoft/WSL/issues/423#issuecomment-608237689
// also based on https://github.com/microsoft/vscode/blob/90a39ba0d49d75e9a4d7e62a6121ad946ecebc58/resources/win32/bin/code.sh#L24
func IsMicrosoftWSL() bool {
	return os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSLPATH") != ""
}

// IsOnGCE determines whether minikube is currently running on GCE.
func IsOnGCE() bool {
	resp, err := http.Get("http://metadata.google.internal")
	if err != nil {
		return false
	}

	return resp.Header.Get("Metadata-Flavor") == "Google"
}

// IsCloudShell determines whether minikube is running inside CloudShell
func IsCloudShell() bool {
	e := os.Getenv("CLOUD_SHELL")
	return e == "true"
}

// IsAmd64M1Emulation  determines whether amd64 minikube binary is running on M1 mac in emulation mode
func IsAmd64M1Emulation() bool {
	return runtime.GOARCH == "amd64" && strings.HasPrefix(cpuid.CPU.BrandName, "VirtualApple")
}

// EffectiveArch return architecture to use in minikube VM/container
// may differ from host arch
func EffectiveArch() string {
	if IsAmd64M1Emulation() {
		return "arm64"
	}
	return runtime.GOARCH
}

// MinikubeInstalledViaSnap returns true if the minikube binary path includes "snap".
func MinikubeInstalledViaSnap() bool {
	ex, err := os.Executable()
	if err != nil {
		return false
	}
	exPath := filepath.Dir(ex)

	return strings.Contains(exPath, "snap")
}

// DockerInstalledViaSnap returns true if the Docker binary path includes "snap".
func DockerInstalledViaSnap() bool {
	c := exec.Command("which", "docker")
	o, err := c.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(o), "snap")
}

// GithubActionRunner returns true if running inside a github action runner
func GithubActionRunner() bool {
	// based on https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

// ImageCacheDir returns the path in the minikube home directory to the container image cache for the current architecture
func ImageCacheDir() string {
	return filepath.Join(localpath.MakeMiniPath("cache", "images"), runtime.GOARCH)
}

// KICCacheDir returns the path in the minikube home directory to the container node cache for the current architecture
func KICCacheDir() string {
	return filepath.Join(localpath.MakeMiniPath("cache", "kic"), runtime.GOARCH)
}

// ISOCacheDir returns the path in the minikube home directory to the virtual machine image cache for the current architecture
func ISOCacheDir() string {
	return filepath.Join(localpath.MakeMiniPath("cache", "iso"), runtime.GOARCH)
}

// SocketVMNetInstalled returns if socket_vmnet is installed
func SocketVMNetInstalled() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := os.Stat(viper.GetString("socket-vmnet-path"))
	if err == nil {
		return true
	}
	if !errors.Is(err, os.ErrNotExist) {
		klog.Warningf("failed to check for socket_vmnet: %v", err)
	}
	return false
}
