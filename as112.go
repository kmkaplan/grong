/* A name server for the AS112 sink system. See http://www.as112.net/

Stephane Bortzmeyer <stephane+grong@bortzmeyer.org>

*/

package responder

import (
	"regexp"
	"strings"
	"container/vector"
	"./types"
)

const as112Regexp = "(168\\.192\\.in-addr\\.arpa|154\\.169\\.in-addr\\.arpa|16\\.172\\.in-addr\\.arpa|17\\.172\\.in-addr\\.arpa|18\\.172\\.in-addr\\.arpa|19\\.172\\.in-addr\\.arpa|20\\.172\\.in-addr\\.arpa|21\\.172\\.in-addr\\.arpa|22\\.172\\.in-addr\\.arpa|23\\.172\\.in-addr\\.arpa|24\\.172\\.in-addr\\.arpa|25\\.172\\.in-addr\\.arpa|26\\.172\\.in-addr\\.arpa|27\\.172\\.in-addr\\.arpa|28\\.172\\.in-addr\\.arpa|29\\.172\\.in-addr\\.arpa|30\\.172\\.in-addr\\.arpa|31\\.172\\.in-addr\\.arpa|10\\.in-addr\\.arpa)$"

const defaultTTL = 3600

var (
	as112Domain    = regexp.MustCompile("^" + as112Regexp)
	as112SubDomain = regexp.MustCompile("\\." + as112Regexp)
	// Answers to "TXT hostname.as112.net"
	hostnameAnswers = [...]string{
		"Unknown location on Earth.",
		"GRONG, name server written in Go.",
		"See http://as112.net/ for more information.",
	}

	// Name servers of AS112, currently two
	as112nameServers = [...]string{
		"blackhole-1.iana.org",
		"blackhole-2.iana.org",
	}

	hostnamesoa = types.SOArecord{
		Mname: "NOT-CONFIGURED.as112.example.net", // Put the real host name
		Rname: "UNKNOWN.as112.example.net",        // Put your email address (with @ replaced by .)
		Serial: 2003030100,
		Refresh: 3600,
		Retry: 600,
		Expire: 2592000,
		Minimum: 15,
	}

	as112soa = types.SOArecord{
		Mname: "prisoner.iana.org",
		Rname: "hostmaster.root-servers.org",
		Serial: 2002040800,
		Refresh: 1800,
		Retry: 900,
		Expire: 604800,
		Minimum: 604800,
	}
)

func nsRecords(domain string, asection *vector.Vector) {
	for i := 0; i < len(as112nameServers); i++ {
		asection.Push(types.RR{
			Name: domain,
			TTL: defaultTTL,
			Type: types.NS,
			Class: types.IN,
			Data: types.Encode(as112nameServers[i]),
		})
	}
}

func soaRecord(domain string, soa types.SOArecord) (result types.RR) {
	result = types.RR{
		Name: domain,
		TTL: defaultTTL,
		Type: types.SOA,
		Class: types.IN,
		Data: types.EncodeSOA(soa),
	}
	return
}

func Respond(query types.DNSquery) (result types.DNSresponse) {
	result.Responsecode = types.SERVFAIL
	qname := strings.ToLower(query.Qname)
	var asection, nssection vector.Vector
	if query.Qclass == types.IN {
		if as112Domain.MatchString(qname) {
			result.Responsecode = types.NOERROR
			if (types.QMatches(query.Qtype, types.NS)) {
				nsRecords(query.Qname, &asection)
			}
			if (types.QMatches(query.Qtype, types.SOA)) {
				asection.Push(soaRecord(query.Qname, as112soa))
			}
		}
		matches := as112SubDomain.MatchStrings(qname)
		if len(matches) > 0 {
			result.Responsecode = types.NXDOMAIN
			nssection.Push(soaRecord(matches[1], as112soa))
		}
		if qname == "hostname.as112.net" {
			result.Responsecode = types.NOERROR
			if types.QMatches(query.Qtype, types.TXT) {
				for i := 0; i < len(hostnameAnswers); i++ {
					asection.Push(types.RR{
						Name: query.Qname,
						TTL: defaultTTL,
						Type: types.TXT,
						Class: types.IN,
						Data: types.ToTXT(hostnameAnswers[i]),
					})
				}
			}
			if (types.QMatches(query.Qtype, types.NS)) {
				nsRecords(query.Qname, &asection)
			}
			if (types.QMatches(query.Qtype, types.SOA)) {
				asection.Push(soaRecord(query.Qname, hostnamesoa))
			}
		}
	}
	data := asection.Data()
	result.Asection = make([]types.RR, len(data))
	for i := 0; i < len(data); i++  {
		result.Asection[i] = data[i].(types.RR)
	}
	data = nssection.Data()
	result.Nssection = make([]types.RR, len(data))
	for i := 0; i < len(data); i++ {
		result.Nssection[i] = data[i].(types.RR)
	}
	return
}
