package x509

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"reconnaissance-service/internal/library/middleware"

	"reconnaissance-service/internal/library/server"
)

func handle(w http.ResponseWriter, r *http.Request) {
	const name = "x509"

	ctx := r.Context()

	labeler, _ := otelhttp.LabelerFromContext(ctx)
	service := middleware.New().Service().Value(ctx)
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(service).Start(ctx, name)

	defer span.End()

	var input Body
	if validator, e := server.Validate(ctx, v, r.Body, &input); e != nil {
		slog.WarnContext(ctx, "Unable to Verify Request Body")

		if validator != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(validator)

			return
		}

		http.Error(w, "Unable to Validate Request Body", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(ctx, "Input", slog.Any("body", input))

	// Construct the address variable that's to be used for generating the tls connection
	address := fmt.Sprintf("%s:%d", input.Hostname, input.Port)

	// Check if hostname exists

	// Attempt to resolve the hostname to an IP address
	_, e := net.LookupIP(input.Hostname)
	if e != nil {
		labeler.Add(attribute.Bool("warning", true))
		slog.WarnContext(ctx, "Hostname Doesn't Exist or Cannot be Resolved", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	// Establish a TLS connection to the hostname
	connection, e := tls.Dial("tcp", address, nil)
	if e != nil {
		labeler.Add(attribute.Bool("error", true))
		slog.ErrorContext(ctx, "Unable to Establish TLS Connection", slog.String("error", e.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Defer closing of the TLS connection, but with a wrapper for unexpected error-handling purposes
	defer func(connection *tls.Conn) {
		e := connection.Close()
		if e != nil {
			labeler.Add(attribute.Bool("warning", true))
			slog.WarnContext(ctx, "Unable to Close TLS Connection", slog.String("error", e.Error()))
		}
	}(connection)

	// Retrieve the TLS connection state
	state := connection.ConnectionState()

	// Check if any certificates are present
	if len(state.PeerCertificates) == 0 {
		message := fmt.Sprintf("No Peer Certificates Found for Address %s", address)

		labeler.Add(attribute.Bool("warning", true))
		slog.WarnContext(ctx, "No Peer Certificates Found", slog.String("address", address), slog.String("message", message))
		http.Error(w, message, http.StatusUnprocessableEntity)
		return
	}

	// Get the leaf certificate (the server's certificate)
	certificate := state.PeerCertificates[0]

	ipaddresses := make([]map[string]interface{}, len(certificate.IPAddresses))
	for index, entity := range certificate.IPAddresses {
		ipaddresses[index] = map[string]interface{}{
			"ipv4":         entity.To4(),
			"ipv6":         entity.To16(),
			"default-mask": entity.DefaultMask().String(),
			"metadata": map[string]bool{
				"global-unicast":            entity.IsGlobalUnicast(),
				"interface-local-multicast": entity.IsInterfaceLocalMulticast(),
				"link-local-multicast":      entity.IsLinkLocalMulticast(),
				"link-local-unicast":        entity.IsLinkLocalUnicast(),
				"loopback":                  entity.IsLoopback(),
				"multicast":                 entity.IsMulticast(),
				"private":                   entity.IsPrivate(),
				"unspecified":               entity.IsUnspecified(),
			},
		}
	}

	exclusions := map[string]interface{}{
		"domains":         certificate.ExcludedDNSDomains,
		"email-addresses": certificate.ExcludedEmailAddresses,
		"ip-ranges":       make([]map[string]interface{}, len(certificate.ExcludedIPRanges)),
		"uri-domains":     certificate.ExcludedURIDomains,
	}

	for index, entity := range certificate.ExcludedIPRanges {
		exclusions["ip-ranges"].([]map[string]interface{})[index] = map[string]interface{}{
			"ipv4":         entity.IP.To4(),
			"ipv6":         entity.IP.To16(),
			"default-mask": entity.IP.DefaultMask().String(),
			"mask":         entity.Mask,
			"network":      entity.Network(),
			"string":       entity.String(),
			"metadata": map[string]bool{
				"global-unicast":            entity.IP.IsGlobalUnicast(),
				"interface-local-multicast": entity.IP.IsInterfaceLocalMulticast(),
				"link-local-multicast":      entity.IP.IsLinkLocalMulticast(),
				"link-local-unicast":        entity.IP.IsLinkLocalUnicast(),
				"loopback":                  entity.IP.IsLoopback(),
				"multicast":                 entity.IP.IsMulticast(),
				"private":                   entity.IP.IsPrivate(),
				"unspecified":               entity.IP.IsUnspecified(),
			},
		}
	}

	permissions := map[string]interface{}{
		"domains":         certificate.PermittedDNSDomains,
		"email-addresses": certificate.PermittedEmailAddresses,
		"ip-ranges":       make([]map[string]interface{}, len(certificate.PermittedIPRanges)),
		"uri-domains":     certificate.PermittedURIDomains,
	}

	for index, entity := range certificate.PermittedIPRanges {
		permissions["ip-ranges"].([]map[string]interface{})[index] = map[string]interface{}{
			"ipv4":         entity.IP.To4(),
			"ipv6":         entity.IP.To16(),
			"default-mask": entity.IP.DefaultMask().String(),
			"mask":         entity.Mask,
			"network":      entity.Network(),
			"string":       entity.String(),
			"metadata": map[string]bool{
				"global-unicast":            entity.IP.IsGlobalUnicast(),
				"interface-local-multicast": entity.IP.IsInterfaceLocalMulticast(),
				"link-local-multicast":      entity.IP.IsLinkLocalMulticast(),
				"link-local-unicast":        entity.IP.IsLinkLocalUnicast(),
				"loopback":                  entity.IP.IsLoopback(),
				"multicast":                 entity.IP.IsMulticast(),
				"private":                   entity.IP.IsPrivate(),
				"unspecified":               entity.IP.IsUnspecified(),
			},
		}
	}

	extensions := make([]map[string]interface{}, len(certificate.Extensions))
	for index, entity := range certificate.Extensions {
		extensions[index] = map[string]interface{}{
			"id":       entity.Id.String(),
			"critical": entity.Critical,
			// @TODO Pro-Version "value":    entity.Value,

		}
	}

	extraextensions := make([]map[string]interface{}, len(certificate.ExtraExtensions))
	for index, entity := range certificate.ExtraExtensions {
		extraextensions[index] = map[string]interface{}{
			"id":       entity.Id.String(),
			"critical": entity.Critical,
			// @TODO Pro-Version "value":    entity.Value,
		}
	}

	extendedextensions := make([]int, len(certificate.ExtKeyUsage))
	for index, entity := range certificate.ExtKeyUsage {
		extendedextensions[index] = int(entity)
	}

	extkeyusages := make([]int, len(certificate.ExtKeyUsage))
	for index, entity := range certificate.ExtKeyUsage {
		extkeyusages[index] = int(entity)
	}

	unhandledcriticalextensions := make([]string, len(certificate.UnhandledCriticalExtensions))
	for index, entity := range certificate.UnhandledCriticalExtensions {
		unhandledcriticalextensions[index] = entity.String()
	}

	unknownextendedkeyusage := make([]string, len(certificate.UnknownExtKeyUsage))
	for index, entity := range certificate.UnhandledCriticalExtensions {
		unknownextendedkeyusage[index] = entity.String()
	}

	issuer := map[string]interface{}{
		"$":                   certificate.Issuer.String(),
		"common-name":         certificate.Issuer.CommonName,
		"country":             certificate.Issuer.Country,
		"locality":            certificate.Issuer.Locality,
		"organization":        certificate.Issuer.Organization,
		"organizational-unit": certificate.Issuer.OrganizationalUnit,
		"postal-code":         certificate.Issuer.PostalCode,
		"province":            certificate.Issuer.Province,
		"street-address":      certificate.Issuer.StreetAddress,
		"serial-number":       certificate.Issuer.SerialNumber,
		"extra-names":         make([]map[string]interface{}, len(certificate.Issuer.ExtraNames)),
		"names":               make([]map[string]interface{}, len(certificate.Issuer.Names)),
	}

	for index, entity := range certificate.Issuer.Names {
		issuer["names"].([]map[string]interface{})[index] = map[string]interface{}{
			"type":  entity.Type.String(),
			"value": entity.Value,
		}
	}

	for index, entity := range certificate.Issuer.ExtraNames {
		issuer["extra-names"].([]map[string]interface{})[index] = map[string]interface{}{
			"type":  entity.Type.String(),
			"value": entity.Value,
		}
	}

	subject := map[string]interface{}{
		"$":                   certificate.Subject.String(),
		"common-name":         certificate.Subject.CommonName,
		"country":             certificate.Subject.Country,
		"locality":            certificate.Subject.Locality,
		"organization":        certificate.Subject.Organization,
		"organizational-unit": certificate.Subject.OrganizationalUnit,
		"postal-code":         certificate.Subject.PostalCode,
		"province":            certificate.Subject.Province,
		"street-address":      certificate.Subject.StreetAddress,
		"serial-number":       certificate.Subject.SerialNumber,
		"extra-names":         make([]map[string]interface{}, len(certificate.Subject.ExtraNames)),
		"names":               make([]map[string]interface{}, len(certificate.Subject.Names)),
	}

	for index, entity := range certificate.Subject.Names {
		subject["names"].([]map[string]interface{})[index] = map[string]interface{}{
			"type":  entity.Type.String(),
			"value": entity.Value,
		}
	}

	for index, entity := range certificate.Subject.ExtraNames {
		subject["names"].([]map[string]interface{})[index] = map[string]interface{}{
			"type":  entity.Type.String(),
			"value": entity.Value,
		}
	}

	policies := make([]string, len(certificate.Policies))
	for index, entity := range certificate.Policies {
		policies[index] = entity.String()
	}

	policyids := make([]string, len(certificate.PolicyIdentifiers))
	for index, entity := range certificate.PolicyIdentifiers {
		policyids[index] = entity.String()
	}

	uris := make([]map[string]interface{}, len(certificate.URIs))
	for index, entity := range certificate.URIs {
		uris[index] = map[string]interface{}{
			"$":                entity.String(),
			"fragment":         entity.Fragment,
			"raw-fragment":     entity.RawFragment,
			"escaped-fragment": entity.EscapedFragment(),
			"path":             entity.Path,
			"raw-path":         entity.RawPath,
			"escaped-path":     entity.EscapedPath(),
			"query":            entity.Query,
			"raw-query":        entity.RawQuery,
			"host":             entity.Host,
			"hostname":         entity.Hostname(),
			"absolute":         entity.IsAbs(),
			"port":             entity.Port(),
			"request-uri":      entity.RequestURI(),
			"scheme":           entity.Scheme,
			"user":             entity.User,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"dns-names":                      certificate.DNSNames,
		"email-addresses":                certificate.EmailAddresses,
		"ip-addresses":                   ipaddresses,
		"exclusions":                     exclusions,
		"authority-key-id":               certificate.AuthorityKeyId,
		"basic-constraints-valid":        certificate.BasicConstraintsValid,
		"crl-distribution-points":        certificate.CRLDistributionPoints,
		"extensions":                     extensions,
		"extra-extensions":               extraextensions,
		"extended-extensions":            extendedextensions,
		"extended-key-usages":            extkeyusages,
		"certificate-authority":          certificate.IsCA,
		"issuer":                         issuer,
		"issuing-certificate-url":        certificate.IssuingCertificateURL,
		"key-usage":                      int(certificate.KeyUsage),
		"max-path-length":                certificate.MaxPathLen,
		"max-path-length-zero":           certificate.MaxPathLenZero,
		"permitted-dns-names":            certificate.PermittedDNSDomains,
		"not-after":                      certificate.NotAfter.String(),
		"not-before":                     certificate.NotBefore.String(),
		"ocsp-server":                    certificate.OCSPServer,
		"permitted-dns-domains-critical": certificate.PermittedDNSDomainsCritical,
		"permissions":                    permissions,
		"policies":                       policies,
		"policy-identifiers":             policyids,
		"public-key":                     certificate.PublicKey,
		"public-key-algorithm":           certificate.PublicKeyAlgorithm.String(),
		"serial-number": map[string]interface{}{
			"$":      certificate.SerialNumber,
			"string": certificate.SerialNumber.String(),
		},
		"signature":                          certificate.Signature,
		"signature-algorithm":                certificate.SignatureAlgorithm.String(),
		"subject":                            subject,
		"subject-key-id":                     certificate.SubjectKeyId,
		"unhandled-critical-extensions":      unhandledcriticalextensions,
		"unknown-extended-key-usage":         unknownextendedkeyusage,
		"uris":                               uris,
		"version":                            certificate.Version,
		"raw":                                certificate.Raw,
		"raw-issuer":                         certificate.RawIssuer,
		"raw-subject":                        certificate.RawSubject,
		"raw-subject-public-key-information": certificate.RawSubjectPublicKeyInfo,
		"raw-tbs-certificate":                certificate.RawTBSCertificate,
	})

	return
}

var Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	handle(w, r)

	return
})
