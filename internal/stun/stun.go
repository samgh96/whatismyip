package stun

import (
        "crypto/rand"
        "fmt"
	"encoding/hex"
)

const (
        MagicCookie = 0x2112a442
        HalfCookie = uint16(MagicCookie >> 16)
)

type StunHeader struct {
        MessageCmd uint16 // StunMessage includes the first two zeroes in the header
        MessageLen uint16
        MagicCookie uint32
        TransactionId []byte
	NetworkVersion uint16
	Patata []byte
}

type StunRequest = StunHeader // StunRequests are only headers TODO: check if this is true

type StunResponse struct {
        Header StunHeader
        AttributeType uint16
        AttributeLen uint16
        AttributeValue []byte
}

func PrintStunRequest(req *StunRequest) {
	fmt.Printf("cmd: %x len: %d cookie: %x tid: %x net: %x\n",
		req.MessageCmd,
		req.MessageLen,
		req.MagicCookie,
		req.TransactionId,
		req.NetworkVersion,
		// req.Patata,
	)
}

func PrintStunResponse(resp *StunResponse) {
	PrintStunRequest(&resp.Header)
	fmt.Printf("attr type: %x attr len: %d attr val: %x\n",
		resp.AttributeType,
		resp.AttributeLen,
		resp.AttributeValue,
	)
}

func XORPort(xored uint16) uint16 {
	return xored ^ HalfCookie
}

func XORIP(xored uint32) uint32 {
	return xored ^ MagicCookie
}

func CreateStunRequest() StunRequest {
        tid := make([]byte, 12)
	pat, _ := hex.DecodeString("000400000000")

        _, err := rand.Read(tid)
        if err != nil {
                fmt.Println("error: ", err)
        }
	
        return StunRequest{
                MessageCmd: 0x001,
                MessageLen: 8,
                MagicCookie: MagicCookie,
                TransactionId: tid,
		NetworkVersion: 0x0003,
		Patata: pat,
        };
}
