package ftpconnection_test

import "net"

var remoteAddress = &net.TCPAddr{
	IP:   net.IP("10.0.0.1"),
	Port: 21,
	Zone: "",
}
