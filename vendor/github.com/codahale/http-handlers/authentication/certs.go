package authentication

import (
	"crypto/x509/pkix"
	"log"
	"net/http"
	"regexp"
)

// X509NameVerifier supports wrapping an http.Handler to check the contents
// of an x509 distinguished name (DN) passed in a header as from Nginx
type X509NameVerifier struct {
	CheckCertificate func(*pkix.Name) bool
	InvalidHandler   http.Handler
	HeaderName       string
}

var dnRegexp *regexp.Regexp

// Wrap wraps an HTTP handler to check the contents of client certificates.
// If CheckCertificate returns true, the request will be passed to the wrapped
// handler. If CheckCertificate returns false, it will be passed to the
// InvalidHandler or, if no InvalidHandler is specified, will return an
// empty 403 response and log the rejected DN.
func (v *X509NameVerifier) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dn := r.Header.Get(v.HeaderName)

		var name *pkix.Name
		if dn != "" {
			name = parseDN(dn)
		}

		if name != nil && v.CheckCertificate(name) {
			h.ServeHTTP(w, r)
		} else if v.InvalidHandler != nil {
			v.InvalidHandler.ServeHTTP(w, r)
		} else {
			log.Printf("Rejected request with an invalid client certificate: %q", dn)
			w.WriteHeader(403)
		}
	})
}

func parseDN(dn string) *pkix.Name {
	name := pkix.Name{}

	matches := dnRegexp.FindAllStringSubmatch(dn, -1)

	for _, match := range matches {
		val := match[2]
		if val == "" {
			continue
		}

		switch match[1] {
		case "C":
			name.Country = append(name.Country, val)
		case "O":
			name.Organization = append(name.Organization, val)
		case "OU":
			name.OrganizationalUnit = append(name.OrganizationalUnit, val)
		case "L":
			name.Locality = append(name.Locality, val)
		case "ST":
			name.Province = append(name.Province, val)
		case "SN":
			name.SerialNumber = val
		case "CN":
			name.CommonName = val
		}
	}

	return &name
}

// Return a CheckCertificate function that returns true IFF one of the
// certificates in the list has an OrganiziationUnit exactly matching
// one of the ones allowed.
var RequireOU = func(allowed []string) func(name *pkix.Name) bool {
	return func(name *pkix.Name) bool {
		for _, haveOU := range name.OrganizationalUnit {
			for _, wantOU := range allowed {
				if haveOU == wantOU {
					return true
				}
			}
		}
		return false
	}
}

func init() {
	dnRegexp = regexp.MustCompile(`[/;,]([^=]+)=([^/;,]+)`)
}
