package client

import (
        "net"
        "fmt"
        "whatismyip/internal/stun"
        "encoding/binary"
        "bytes"
        "strings"
        "strconv"
        "time"
        //      "encoding/gob"
)


const STUNServer = "stun.l.google.com:19302"

func UDPServer(local_ip string, local_port int) stun.StunResponse {
        srv, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(local_ip), Port: local_port})
        if err != nil {
                fmt.Println(err)
        }
        defer srv.Close()
	
        buf := make([]byte, 32) // 128 bytes seems reasonable

        // remote_expected_ip, err := net.LookupIP(STUNServer)

        for {
                // n, addr, err := srv.ReadFrom(buf)
		n, _, err := srv.ReadFrom(buf)
                if err != nil {
                        fmt.Println(err)
                }
                // fmt.Println(addr)
                // if addr.String() != remote_expected_ip {
                //         fmt.Println("mismatch: expected " + STUNServer + " but got " + addr.String())
                // }

                if n != 0 {
                        // fmt.Printf("%x\n", buf)
                        return stun.StunResponse{
                                Header: stun.StunHeader{
                                        MessageCmd: binary.BigEndian.Uint16(buf),
                                        MessageLen: binary.BigEndian.Uint16(buf[2:4]),
                                        MagicCookie: binary.BigEndian.Uint32(buf[4:8]),
                                        TransactionId: buf[8:20],
                                        NetworkVersion: binary.BigEndian.Uint16(buf[20:22]),
                                        Patata: nil,
                                },
                                AttributeType: binary.BigEndian.Uint16(buf[22:24]),
                                AttributeLen: binary.BigEndian.Uint16(buf[24:26]),
                                AttributeValue: buf[26:n],
                        }
                }
        }
}

func UDPClient() {
        conn, err := net.DialTimeout("udp4", STUNServer, 3 * time.Second) // 3 seconds timeout
        if err != nil {
                fmt.Println(err)
                return
        }

        local_addr := conn.LocalAddr().String()
        local_ip := strings.Split(local_addr, ":")[0]
        local_port, _ := strconv.Atoi(strings.Split(local_addr, ":")[1])

        // fmt.Println("local address: " + local_addr)
        // fmt.Println("local IP: " + local_ip)
        // fmt.Printf("local port: %d\n", local_port)

        msg := stun.CreateStunRequest()
        // stun.PrintStunRequest(&msg)
        buf := new(bytes.Buffer)

        binary.Write(buf, binary.BigEndian, msg.MessageCmd)
        binary.Write(buf, binary.BigEndian, msg.MessageLen)
        binary.Write(buf, binary.BigEndian, msg.MagicCookie)
        binary.Write(buf, binary.BigEndian, msg.TransactionId)
        binary.Write(buf, binary.BigEndian, msg.NetworkVersion)
        binary.Write(buf, binary.BigEndian, msg.Patata)

        // fmt.Printf("buf len: %d\n", buf.Len())
        // fmt.Printf("%x\n", buf.Bytes())

        conn.Write(buf.Bytes())
        conn.Close()

        response := UDPServer(local_ip, local_port)
        // stun.PrintStunResponse(&response)
	
	fmt.Printf("port: %d\n", stun.XORPort(binary.BigEndian.Uint16(response.AttributeValue[0:2])))

	num_IP := stun.XORIP(binary.BigEndian.Uint32(response.AttributeValue[2:]))
	response_IP := make(net.IP, 4)
	binary.BigEndian.PutUint32(response_IP, num_IP)
	fmt.Printf("public IP: %s\n", response_IP)
	
}
