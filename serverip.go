package serverip

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"

	"github.com/vishvananda/netlink"
)

const name = "serverip"

type Serverip struct{}

func getSourceIP(destinationIP string) (net.IP, error) {
	// Parse the destination IP address
	dstIP := net.ParseIP(destinationIP)
	if dstIP == nil {
		return nil, fmt.Errorf("invalid destination IP address")
	}

	// Lookup route to the destination IP
	routes, err := netlink.RouteGet(dstIP)
	if err != nil {
		return nil, fmt.Errorf("error looking up route: %v", err)
	}

	return routes[0].Src, nil
}

// ServeDNS implements the plugin.Handler interface.
func (si Serverip) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true

	ip, err := getSourceIP(state.IP())
	if err != nil {
		return 1, err
	}

	var rr dns.RR

	switch state.Family() {
	case 1:
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass()}
		rr.(*dns.A).A = ip.To4()
	case 2:
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: state.QClass()}
		rr.(*dns.AAAA).AAAA = ip
	}

	srv := new(dns.SRV)
	srv.Hdr = dns.RR_Header{Name: "_" + state.Proto() + "." + state.QName(), Rrtype: dns.TypeSRV, Class: state.QClass()}
	if state.QName() == "." {
		srv.Hdr.Name = "_" + state.Proto() + state.QName()
	}
	port, _ := strconv.ParseUint(state.Port(), 10, 16)
	srv.Port = uint16(port)
	srv.Target = "."

	a.Extra = []dns.RR{rr, srv}

	w.WriteMsg(a)

	return 0, nil
}

// Name implements the Handler interface.
func (si Serverip) Name() string { return name }
