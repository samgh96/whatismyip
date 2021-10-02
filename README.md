# whatsmyip
## Disclaimer

This is a WIP! In order to make it work I've been cutting some corners: the
code is not as pretty as it could be and also the docs are not up to date
with how I approached this and found out how it works.

## Usage

Run `go run cmd/whatismyip/whatismyip.go`

## Concept

Instead of relying in a two-side deployment (one bot in your local network
and one in a server you own) this bot uses WebRTC to get its public address.

`whatsmyip` sends a packet to a public STUN server. You could use your 
own STUN server or rely on one of the included server URLs, that's up to you.

### STUN request/response

From [https://datatracker.ietf.org/doc/html/rfc8489#section-5](RFC 8489)

```
   All STUN messages comprise a 20-byte header followed by zero or more
   attributes.  The STUN header contains a STUN message type, message
   length, magic cookie, and transaction ID.

      0                   1                   2                   3
      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |0 0|     STUN Message Type     |         Message Length        |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                         Magic Cookie                          |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                                                               |
     |                     Transaction ID (96 bits)                  |
     |                                                               |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

                  Figure 2: Format of STUN Message Header

   The most significant 2 bits of every STUN message MUST be zeroes.
   This can be used to differentiate STUN packets from other protocols
   when STUN is multiplexed with other protocols on the same port.

   The message type defines the message class (request, success
   response, error response, or indication) and the message method (the
   primary function) of the STUN message. [...] The Magic Cookie field 
   MUST contain the fixed value 0x2112A442 in network byte order.
   
```

Since we only want to do a binding request (the response will give us our
public IP) our STUN message will be this one:

```
      0                   1                   2                   3
      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |0 0|   0x0001 (Binding req)    |               8               |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                         0x2112A442                            |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                                                               |
     |                     random(0, 2**96-1)                        |
     |                                                               |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

In order to read what we receive we need to introduce the concept of 
STUN attributes (from [https://datatracker.ietf.org/doc/html/rfc8489#section-14](RFC 8489))

```
   After the STUN header are zero or more attributes.  Each attribute
   MUST be TLV encoded, with a 16-bit type, 16-bit length, and value.
   Each STUN attribute MUST end on a 32-bit boundary.  As mentioned
   above, all fields in an attribute are transmitted most significant
   bit first.

      0                   1                   2                   3
      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |         Type                  |            Length             |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                         Value (variable)                ....
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

                    Figure 4: Format of STUN Attributes

   The value in the Length field MUST contain the length of the Value
   part of the attribute, prior to padding, measured in bytes.  Since
   STUN aligns attributes on 32-bit boundaries, attributes whose content
   is not a multiple of 4 bytes are padded with 1, 2, or 3 bytes of
   padding so that its value contains a multiple of 4 bytes.  The
   padding bits MUST be set to zero on sending and MUST be ignored by
   the receiver.

```

We are receiving an XOR-mapped address:

```
   The XOR-MAPPED-ADDRESS attribute is identical to the MAPPED-ADDRESS
   attribute, except that the reflexive transport address is obfuscated
   through the XOR function.

   The format of the XOR-MAPPED-ADDRESS is:

      0                   1                   2                   3
      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |0 0 0 0 0 0 0 0|    Family     |         X-Port                |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                X-Address (Variable)
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

             Figure 6: Format of XOR-MAPPED-ADDRESS Attribute

   The Family field represents the IP address family and is encoded
   identically to the Family field in MAPPED-ADDRESS, which is:

   0x01:IPv4
   0x02:IPv6

   X-Port is computed by XOR'ing the mapped port with the most
   significant 16 bits of the magic cookie.  If the IP address family is
   IPv4, X-Address is computed by XOR'ing the mapped IP address with the
   magic cookie.  If the IP address family is IPv6, X-Address is
   computed by XOR'ing the mapped IP address with the concatenation of
   the magic cookie and the 96-bit transaction ID.  In all cases, the
   XOR operation works on its inputs in network byte order (that is, the
   order they will be encoded in the message).
```

So in our case the response will be something like this:

```
      0                   1                   2                   3
      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |0 0|   0x0101 (Binding res)    |              12               |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                         0x2112A442                            |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                                                               |
     |                  the same transaction ID we sent              |
     |                                                               |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |0 0 0 0 0 0 0 0|  0x01 (IPv4)  |         X-Port                |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
     |                X-Address (Variable)                           |
     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

For example, if our port is `1337`, we should receive `0x242b`, because
`1337 ^ 0x2112 = 0x242b`. Same with our IP, but instead of using the 16
most significant bytes of the magic cookie we need to use the whole
cookie, so e.g. if our IP is `123.123.8.8`:

 - Our IP in hex would be `0x7b7b0808` (convert each segment to hex).
 - XOR'ing the whole cookie: `0x7b7b0808 ^ 0x2112a442 = 0x5a69ac4a`

## Docs
- https://help.singlecomm.com/hc/en-us/articles/115007993947-STUN-servers-A-Quick-Start-Guide
- https://datatracker.ietf.org/doc/html/rfc8489
