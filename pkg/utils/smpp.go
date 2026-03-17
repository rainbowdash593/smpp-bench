package utils

import (
	"fmt"

	"github.com/fiorix/go-smpp/smpp/pdu"
)

func PDUToString(p pdu.Body) string {
	return fmt.Sprintf(
		"seq=%d, command=%s, status=%s, fields=%+v, tlv=%+v",
		p.Header().Seq, p.Header().ID.String(),
		p.Header().Status.Error(), p.Fields(), p.TLVFields(),
	)
}
