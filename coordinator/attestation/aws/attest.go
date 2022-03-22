//go:build aws
// +build aws

package aws

// #include <nitroattest.h>
import "C"

import (
	"fmt"
	"time"
)

func NaAdGetVerifiedPayloadAsJson(adBlob []byte, rootCertDer []byte, ts time.Time) (string, error) {
	jsonCstr := C.na_ad_get_verified_payload_as_json(
		(*C.uint8_t)(&adBlob[0]),
		C.size_t(len(adBlob)),
		(*C.uint8_t)(&rootCertDer[0]),
		C.size_t(len(rootCertDer)),
		C.uint64_t(ts.Unix()),
	)
	jsonStr := C.GoString(jsonCstr)
	C.na_str_free(jsonCstr)

	if jsonStr == "" {
		return "", fmt.Errorf("failed to verify attestation document: %s", adBlob)
	}

	return jsonStr, nil
}
