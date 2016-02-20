package channels

import "testing"
import "net/url"
import "strings"

func TestVerify(t *testing.T) {
	v := url.Values{}
	v.Set("expire", "1455998487")
	v.Set("realm", "838412")
	v.Set("user", "cesar")
	v.Set("_token", "b6b8e488f17e3da62edc909fc00f2b76a902879c")
	// Now is the expiry minus 10 seconds
	err := validateWsRequest(v, 1455998487-10)
	if err != nil {
		t.Error(err)
	}
}

func TestVerifyBadSignature(t *testing.T) {
	v := url.Values{}
	v.Set("expire", "1455998487")
	v.Set("realm", "838412")
	v.Set("user", "cesar")
	v.Set("_token", "cafebaecafebaecafebaecafebae")
	// Now is the expiry minus 10 seconds
	err := validateWsRequest(v, 1455998487-10)
	t.Log("Got err", err)
	if err == nil || !strings.Contains(err.Error(),
		"signature was not correct") {
		t.Error("Should have gotten an invalid signature")
	}

}

func TestVerifyBadSignatureHex(t *testing.T) {
	v := url.Values{}
	v.Set("expire", "1455998487")
	v.Set("realm", "838412")
	v.Set("user", "cesar")
	v.Set("_token", "foobar")
	err := validateWsRequest(v, 1455998487-10)
	t.Log("Got err", err)
	if err == nil || !strings.Contains(err.Error(),
		"encoding/hex") {
		t.Error("Should have gotten an invalid signature - encoding/hex")
	}

}

func TestVerifyExpired(t *testing.T) {
	v := url.Values{}
	v.Set("expire", "1455998487")
	v.Set("realm", "838412")
	v.Set("user", "cesar")
	v.Set("_token", "b6b8e488f17e3da62edc909fc00f2b76a902879c")
	// We expired by a second!
	err := validateWsRequest(v, 1455998487+1)
	if err == nil || !strings.Contains(err.Error(), "your token has expired") {
		t.Error("Should have gotten a token expiry error")
	}
}

func TestVerifyNoRealm(t *testing.T) {
	v := url.Values{}
	v.Set("expire", "1455998487")
	v.Set("user", "cesar")
	v.Set("_token", "b6b8e488f17e3da62edc909fc00f2b76a902879c")
	// No realm specified (signature is wrong too but that's beside the point)
	err := validateWsRequest(v, 1455998487-10)
	if err == nil || !strings.Contains(err.Error(), "no realm") {
		t.Error("Should have gotten a realm error")
	}
}
