/*
Copyright 2014 The Kubernetes Authors.

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

package validation

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"strings"
)

const qnameCharFmt string = "[A-Za-z0-9]"
const qnameExtCharFmt string = "[-A-Za-z0-9_.]"
const qualifiedNameFmt string = "(" + qnameCharFmt + qnameExtCharFmt + "*)?" + qnameCharFmt
const qualifiedNameMaxLength int = 63

var qualifiedNameRegexp = regexp.MustCompile("^" + qualifiedNameFmt + "$")

// IsQualifiedName tests whether the value passed is what Kubernetes calls a
// "qualified name".  This is a format used in various places throughout the
// system.  If the value is not valid, a list of error strings is returned.
// Otherwise an empty list (or nil) is returned.
func IsQualifiedName(value string) []string {
	var errs []string
	parts := strings.Split(value, "/")
	var name string
	switch len(parts) {
	case 1:
		name = parts[0]
	case 2:
		var prefix string
		prefix, name = parts[0], parts[1]
		if len(prefix) == 0 {
			errs = append(errs, "prefix part "+EmptyError())
		} else if msgs := IsDNS1123Subdomain(prefix); len(msgs) != 0 {
			errs = append(errs, prefixEach(msgs, "prefix part ")...)
		}
	default:
		return append(errs, RegexError(qualifiedNameFmt, "MyName", "my.name", "123-abc")+
			" with an optional DNS subdomain prefix and '/' (e.g. 'example.com/MyName'")
	}

	if len(name) == 0 {
		errs = append(errs, "name part "+EmptyError())
	} else if len(name) > qualifiedNameMaxLength {
		errs = append(errs, "name part "+MaxLenError(qualifiedNameMaxLength))
	}
	if !qualifiedNameRegexp.MatchString(name) {
		errs = append(errs, "name part "+RegexError(qualifiedNameFmt, "MyName", "my.name", "123-abc"))
	}
	return errs
}

const labelValueFmt string = "(" + qualifiedNameFmt + ")?"
const LabelValueMaxLength int = 63

var labelValueRegexp = regexp.MustCompile("^" + labelValueFmt + "$")

// IsValidLabelValue tests whether the value passed is a valid label value.  If
// the value is not valid, a list of error strings is returned.  Otherwise an
// empty list (or nil) is returned.
func IsValidLabelValue(value string) []string {
	var errs []string
	if len(value) > LabelValueMaxLength {
		errs = append(errs, MaxLenError(LabelValueMaxLength))
	}
	if !labelValueRegexp.MatchString(value) {
		errs = append(errs, RegexError(labelValueFmt, "MyValue", "my_value", "12345"))
	}
	return errs
}

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const DNS1123LabelMaxLength int = 63

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

// IsDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func IsDNS1123Label(value string) []string {
	var errs []string
	if len(value) > DNS1123LabelMaxLength {
		errs = append(errs, MaxLenError(DNS1123LabelMaxLength))
	}
	if !dns1123LabelRegexp.MatchString(value) {
		errs = append(errs, RegexError(dns1123LabelFmt, "my-name", "123-abc"))
	}
	return errs
}

const dns1123SubdomainFmt string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"
const DNS1123SubdomainMaxLength int = 253

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

// IsDNS1123Subdomain tests for a string that conforms to the definition of a
// subdomain in DNS (RFC 1123).
func IsDNS1123Subdomain(value string) []string {
	var errs []string
	if len(value) > DNS1123SubdomainMaxLength {
		errs = append(errs, MaxLenError(DNS1123SubdomainMaxLength))
	}
	if !dns1123SubdomainRegexp.MatchString(value) {
		errs = append(errs, RegexError(dns1123SubdomainFmt, "example.com"))
	}
	return errs
}

const dns1035LabelFmt string = "[a-z]([-a-z0-9]*[a-z0-9])?"
const DNS1035LabelMaxLength int = 63

var dns1035LabelRegexp = regexp.MustCompile("^" + dns1035LabelFmt + "$")

// IsDNS1035Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1035).
func IsDNS1035Label(value string) []string {
	var errs []string
	if len(value) > DNS1035LabelMaxLength {
		errs = append(errs, MaxLenError(DNS1035LabelMaxLength))
	}
	if !dns1035LabelRegexp.MatchString(value) {
		errs = append(errs, RegexError(dns1035LabelFmt, "my-name", "abc-123"))
	}
	return errs
}

// wildcard definition - RFC 1034 section 4.3.3.
// examples:
// - valid: *.bar.com, *.foo.bar.com
// - invalid: *.*.bar.com, *.foo.*.com, *bar.com, f*.bar.com, *
const wildcardDNF1123SubdomainFmt = "\\*\\." + dns1123SubdomainFmt

// IsWildcardDNS1123Subdomain tests for a string that conforms to the definition of a
// wildcard subdomain in DNS (RFC 1034 section 4.3.3).
func IsWildcardDNS1123Subdomain(value string) []string {
	wildcardDNS1123SubdomainRegexp := regexp.MustCompile("^\\*\\." + dns1123SubdomainFmt + "$")

	var errs []string
	if len(value) > DNS1123SubdomainMaxLength {
		errs = append(errs, MaxLenError(DNS1123SubdomainMaxLength))
	}
	if !wildcardDNS1123SubdomainRegexp.MatchString(value) {
		errs = append(errs, RegexError(wildcardDNF1123SubdomainFmt, "*.example.com"))
	}
	return errs
}

const cIdentifierFmt string = "[A-Za-z_][A-Za-z0-9_]*"

var cIdentifierRegexp = regexp.MustCompile("^" + cIdentifierFmt + "$")

// IsCIdentifier tests for a string that conforms the definition of an identifier
// in C. This checks the format, but not the length.
func IsCIdentifier(value string) []string {
	if !cIdentifierRegexp.MatchString(value) {
		return []string{RegexError(cIdentifierFmt, "my_name", "MY_NAME", "MyName")}
	}
	return nil
}

// IsValidPortNum tests that the argument is a valid, non-zero port number.
func IsValidPortNum(port int) []string {
	if 1 <= port && port <= 65535 {
		return nil
	}
	return []string{InclusiveRangeError(1, 65535)}
}

// Now in libcontainer UID/GID limits is 0 ~ 1<<31 - 1
// TODO: once we have a type for UID/GID we should make these that type.
const (
	minUserID  = 0
	maxUserID  = math.MaxInt32
	minGroupID = 0
	maxGroupID = math.MaxInt32
)

// IsValidGroupId tests that the argument is a valid Unix GID.
func IsValidGroupId(gid int64) []string {
	if minGroupID <= gid && gid <= maxGroupID {
		return nil
	}
	return []string{InclusiveRangeError(minGroupID, maxGroupID)}
}

// IsValidUserId tests that the argument is a valid Unix UID.
func IsValidUserId(uid int64) []string {
	if minUserID <= uid && uid <= maxUserID {
		return nil
	}
	return []string{InclusiveRangeError(minUserID, maxUserID)}
}

var portNameCharsetRegex = regexp.MustCompile("^[-a-z0-9]+$")
var portNameOneLetterRegexp = regexp.MustCompile("[a-z]")

// IsValidPortName check that the argument is valid syntax. It must be
// non-empty and no more than 15 characters long. It may contain only [-a-z0-9]
// and must contain at least one letter [a-z]. It must not start or end with a
// hyphen, nor contain adjacent hyphens.
//
// Note: We only allow lower-case characters, even though RFC 6335 is case
// insensitive.
func IsValidPortName(port string) []string {
	var errs []string
	if len(port) > 15 {
		errs = append(errs, MaxLenError(15))
	}
	if !portNameCharsetRegex.MatchString(port) {
		errs = append(errs, "must contain only alpha-numeric characters (a-z, 0-9), and hyphens (-)")
	}
	if !portNameOneLetterRegexp.MatchString(port) {
		errs = append(errs, "must contain at least one letter (a-z)")
	}
	if strings.Contains(port, "--") {
		errs = append(errs, "must not contain consecutive hyphens")
	}
	if len(port) > 0 && (port[0] == '-' || port[len(port)-1] == '-') {
		errs = append(errs, "must not begin or end with a hyphen")
	}
	return errs
}

// IsValidIP tests that the argument is a valid IP address.
func IsValidIP(value string) []string {
	if net.ParseIP(value) == nil {
		return []string{"must be a valid IP address, (e.g. 10.9.8.7)"}
	}
	return nil
}

const percentFmt string = "[0-9]+%"

var percentRegexp = regexp.MustCompile("^" + percentFmt + "$")

func IsValidPercent(percent string) []string {
	if !percentRegexp.MatchString(percent) {
		return []string{RegexError(percentFmt, "1%", "93%")}
	}
	return nil
}

const httpHeaderNameFmt string = "[-A-Za-z0-9]+"

var httpHeaderNameRegexp = regexp.MustCompile("^" + httpHeaderNameFmt + "$")

// IsHTTPHeaderName checks that a string conforms to the Go HTTP library's
// definition of a valid header field name (a stricter subset than RFC7230).
func IsHTTPHeaderName(value string) []string {
	if !httpHeaderNameRegexp.MatchString(value) {
		return []string{RegexError(httpHeaderNameFmt, "X-Header-Name")}
	}
	return nil
}

const configMapKeyFmt = `[-._a-zA-Z0-9]+`

var configMapKeyRegexp = regexp.MustCompile("^" + configMapKeyFmt + "$")

// IsConfigMapKey tests for a string that is a valid key for a ConfigMap or Secret
func IsConfigMapKey(value string) []string {
	var errs []string
	if len(value) > DNS1123SubdomainMaxLength {
		errs = append(errs, MaxLenError(DNS1123SubdomainMaxLength))
	}
	if !configMapKeyRegexp.MatchString(value) {
		errs = append(errs, RegexError(configMapKeyFmt, "key.name", "KEY_NAME", "key-name"))
	}
	if value == "." {
		errs = append(errs, `must not be '.'`)
	}
	if value == ".." {
		errs = append(errs, `must not be '..'`)
	} else if strings.HasPrefix(value, "..") {
		errs = append(errs, `must not start with '..'`)
	}
	return errs
}

// MaxLenError returns a string explanation of a "string too long" validation
// failure.
func MaxLenError(length int) string {
	return fmt.Sprintf("must be no more than %d characters", length)
}

// RegexError returns a string explanation of a regex validation failure.
func RegexError(fmt string, examples ...string) string {
	s := "must match the regex " + fmt
	if len(examples) == 0 {
		return s
	}
	s += " (e.g. "
	for i := range examples {
		if i > 0 {
			s += " or "
		}
		s += "'" + examples[i] + "'"
	}
	return s + ")"
}

// EmptyError returns a string explanation of a "must not be empty" validation
// failure.
func EmptyError() string {
	return "must be non-empty"
}

func prefixEach(msgs []string, prefix string) []string {
	for i := range msgs {
		msgs[i] = prefix + msgs[i]
	}
	return msgs
}

// InclusiveRangeError returns a string explanation of a numeric "must be
// between" validation failure.
func InclusiveRangeError(lo, hi int) string {
	return fmt.Sprintf(`must be between %d and %d, inclusive`, lo, hi)
}
