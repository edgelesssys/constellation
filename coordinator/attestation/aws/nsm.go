//go:build aws
// +build aws

package aws

// #include <nsm.h>
import "C"

import (
	"fmt"
)

// As defined by the attestation document's COSE_Sign1 structure
const nsmMaxAttestationDocSize = 16 * 1024

func NsmGetAttestationDoc(userData []byte, nonce []byte) ([]byte, error) {
	doc := make([]byte, nsmMaxAttestationDocSize)
	doclen := C.uint32_t(len(doc))

	nsm_fd := C.nsm_lib_init()
	if nsm_fd < 0 {
		return nil, fmt.Errorf("could not open NSM module")
	}
	defer C.nsm_lib_exit(nsm_fd)

	errCode := C.nsm_get_attestation_doc(
		nsm_fd,
		(*C.uint8_t)(&userData[0]),
		C.uint32_t(len(userData)),
		(*C.uint8_t)(&nonce[0]),
		C.uint32_t(len(nonce)),
		nil,
		0,
		(*C.uint8_t)(&doc[0]),
		&doclen,
	)
	if errCode != C.ERROR_CODE_SUCCESS {
		return nil, fmt.Errorf("failed to generate attestation document: %d", errCode)
	}
	doc = doc[:doclen]

	return doc, nil
}
