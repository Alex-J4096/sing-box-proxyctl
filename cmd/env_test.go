package cmd

import (
	"reflect"
	"testing"

	"github.com/Alex-J4096/proxyctl/util"
)

func TestProxyEnvLines(t *testing.T) {
	settings := util.DefaultProxyctlSettings()
	settings.Inbound.Mode = "mixed"
	settings.Inbound.MixedPort = 2080
	want := []string{
		"unset ALL_PROXY HTTP_PROXY HTTPS_PROXY all_proxy http_proxy https_proxy",
		"export ALL_PROXY=socks5h://127.0.0.1:2080 all_proxy=socks5h://127.0.0.1:2080",
		"export HTTP_PROXY=http://127.0.0.1:2080 http_proxy=http://127.0.0.1:2080",
		"export HTTPS_PROXY=http://127.0.0.1:2080 https_proxy=http://127.0.0.1:2080",
	}
	if got := proxyEnvLines(settings); !reflect.DeepEqual(got, want) {
		t.Fatalf("proxyEnvLines() = %q, want %q", got, want)
	}
}

func TestProxyUnsetEnvLinesClearsUpperAndLowerCase(t *testing.T) {
	want := []string{"unset ALL_PROXY HTTP_PROXY HTTPS_PROXY all_proxy http_proxy https_proxy"}
	if got := proxyUnsetEnvLines(); !reflect.DeepEqual(got, want) {
		t.Fatalf("proxyUnsetEnvLines() = %q, want %q", got, want)
	}
}

func TestShellQuote(t *testing.T) {
	if got, want := shellQuote("a'b"), `'a'\''b'`; got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}
