// Package normalize contains functions to normalize usernames, domains and
// addresses.
package normalize

import (
	"strings"

	"blitiri.com.ar/go/chasquid/internal/envelope"
	"golang.org/x/net/idna"
	"golang.org/x/text/secure/precis"
	"golang.org/x/text/unicode/norm"
)

// User normalizes an username using PRECIS.
// On error, it will also return the original username to simplify callers.
func User(user string) (string, error) {
	norm, err := precis.UsernameCaseMapped.String(user)
	if err != nil {
		return user, err
	}

	return norm, nil
}

// Domain normalizes a DNS domain into a cleaned UTF-8 form.
// On error, it will also return the original domain to simplify callers.
func Domain(domain string) (string, error) {
	// For now, we just convert them to lower case and make sure it's in NFC
	// form for consistency.
	// There are other possible transformations (like nameprep) but for our
	// purposes these should be enough.
	// https://tools.ietf.org/html/rfc5891#section-5.2
	// https://blog.golang.org/normalization
	d, err := idna.ToUnicode(domain)
	if err != nil {
		return domain, err
	}

	d = norm.NFC.String(d)
	d = strings.ToLower(d)
	return d, nil
}

// Name normalizes an email address, applying User and Domain to its
// respective components.
// On error, it will also return the original address to simplify callers.
func Addr(addr string) (string, error) {
	user, domain := envelope.Split(addr)

	user, err := User(user)
	if err != nil {
		return addr, err
	}

	domain, err = Domain(domain)
	if err != nil {
		return addr, err
	}

	return user + "@" + domain, nil
}

// Take an address with an ASCII domain, and convert it to Unicode as per
// IDNA, including basic normalization.
// The user part is unchanged.
func DomainToUnicode(addr string) (string, error) {
	if addr == "<>" {
		return addr, nil
	}
	user, domain := envelope.Split(addr)

	domain, err := Domain(domain)
	return user + "@" + domain, err
}