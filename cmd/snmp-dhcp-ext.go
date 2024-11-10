package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type LeaseRaw struct {
	ip             string
	starts         time.Time
	ends           time.Time
	tstp           time.Time
	cltt           time.Time
	binding        string
	next           string
	rewind         string
	hardware       string
	uid            string
	clientHostname string
	hostname       string
}

type Lease struct {
	Ip       string `json:"ip""`
	Fqdn string `json:"fqdn"`
	Hostname string `json:"hostname"`
	Mac      string `json:"mac"`
}

type Subnet struct {
	Network string
	Mask    string
	Domain  string
}

const (
	LeaseStartStatement = "lease"
	LeaseEndWord    = "}"
	SubnetStartStatement = "subnet"
	SubnetEndWord   = "}"
	DhcpdConfPath   = "/var/dhcpd/etc/dhcpd.conf"
	DhcpdLeasesPath = "/var/dhcpd/var/db/dhcpd.leases"
)

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(w io.Writer, _ []string) error {
	dlf, err := os.Open(DhcpdLeasesPath)
	if err != nil {
		return fmt.Errorf("could not open file %s: %s", DhcpdLeasesPath, err)
	}
	dlr := bufio.NewReader(dlf)
	dls := bufio.NewScanner(dlr)
	defer dlf.Close()

	dcf, err := os.Open(DhcpdConfPath)
	if err != nil {
		return fmt.Errorf("could not open file %s: %s", DhcpdConfPath, err)
	}
	dcr := bufio.NewReader(dcf)
	dcs := bufio.NewScanner(dcr)
	defer dcf.Close()

	subnets, err := parseConfigFile(dcs)
	if err != nil {
		return fmt.Errorf("could not parse config file: %s", err)
	}

	rawLeases, err := parseLeaseFile(dls)
	if err != nil {
		return err
	}

	writer := json.NewEncoder(w)
	for _, rl := range(rawLeases) {
		lease := buildLease(rl, subnets)
		if lease != nil {
		    writer.Encode(lease)
		}
	}

	return nil
}

func leaseFilter(l *LeaseRaw) bool {
	if l.binding == "abandoned" {
		return true
	}
	if len(l.hardware) <= 0 {
		return true
	}
	if len(l.hostname) <= 0 && len(l.clientHostname) <= 0 {
		return true
	}

	return false
}

func buildLease(l *LeaseRaw, subnets []*Subnet) *Lease {
	lease := Lease{}

	lease.Ip = l.ip
	lease.Mac = l.hardware

	if len(l.hostname) > 0 {
		lease.Hostname = strings.ToLower(l.hostname)
	}
	if len(l.clientHostname) > 0 {
		lease.Hostname = strings.ToLower(l.clientHostname)
	}

	var match *Subnet
	for _, s := range subnets {
		if subnetContainsIp(l.ip, s) {
			match = s
		}
	}

	if match == nil || len(match.Domain) <= 0 {
		return nil
	}

	lease.Fqdn = fmt.Sprintf("%s.%s", lease.Hostname, match.Domain)

	return &lease
}

func subnetContainsIp(ip string, subnet *Subnet) bool {
	ipOctets := strings.Split(ip, ".")
	netOctets := strings.Split(subnet.Network, ".")
	maskOctets := strings.Split(subnet.Mask, ".")

	for i, _ := range ipOctets {
		// NOTE: only supporting whole octet masks for now
		if strings.Compare(maskOctets[i], "255") != 0 {
			continue
		}

		if strings.Compare(ipOctets[i], netOctets[i]) != 0 {
			return false
		}
	}

	return true
}

func newLeaseRaw() *LeaseRaw {
	return &LeaseRaw{}
}

func newSubnet() *Subnet {
	return &Subnet{}
}

func parseConfigFile(s *bufio.Scanner) ([]*Subnet, error) {
	var subnets []*Subnet

	for {
		subnet, err := parseSubnet(s)
		if err != nil {
			return nil, err
		}

		if subnet == nil {
			break
		}

		subnets = append(subnets, subnet)
	}

	return subnets, nil
}

func parseSubnet(s *bufio.Scanner) (*Subnet, error) {
	inSubnetBlock := false
	subnet := newSubnet()
	var line string

	for s.Scan() {
		line = s.Text()

		if strings.HasPrefix(line, SubnetStartStatement) {
			inSubnetBlock = true
		}
		if !inSubnetBlock {
			continue
		}
		if strings.HasPrefix(line, SubnetEndWord) {
			return subnet, nil
		}
		if strings.Compare(line, "") == 0 {
			continue
		}

		sl := strings.Split(strings.Trim(line, "\t ;"), " ")

		if err := parseSubnetLine(sl, subnet); err != nil {
			return nil, err
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func parseSubnetLine(sl []string, subnet *Subnet) error {
	if len(sl) == 0 {
		return nil
	}

	stmt := sl[0]

	switch stmt {
	case "subnet":
		subnet.Network = sl[1]
		subnet.Mask = sl[3]
	case "option":
		if len(sl) == 3 && sl[0] == "option" && sl[1] == "domain-name" {
			subnet.Domain = strings.Trim(sl[2], "\";")
		}
	case "ping-check":
		// not implemented
	case "pool":
		// not implemented
	case "range":
		// not implemented
	case "}":
		// no-op
	default:
		return fmt.Errorf("skipping subnet statement with no configured parser '%s'\n%s",
			stmt, strings.Join(sl, ","))
	}

	return nil
}

func parseLeaseFile(s *bufio.Scanner) (map[string]*LeaseRaw, error) {
	leases := make(map[string]*LeaseRaw)

	for {
		raw, err := parseLease(s)
		if err != nil {
			return nil, err
		}
		if raw == nil {
			break
		}

		if !leaseFilter(raw) {
			leases[raw.ip] = raw
		}
	}

	return leases, nil
}
func parseLease(s *bufio.Scanner) (*LeaseRaw, error) {
	inLeaseBlock := false
	lease := newLeaseRaw()
	var line string

	for s.Scan() {
		line = s.Text()

		if strings.HasPrefix(line, LeaseStartStatement) {
			inLeaseBlock = true
		}
		if !inLeaseBlock {
			continue
		}
		if strings.HasPrefix(line, LeaseEndWord) {
			return lease, nil
		}

		ll := strings.Split(strings.Trim(line, " "), " ")
		if err := parseLeaseLine(ll, lease); err != nil {
			return nil, err
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}

func parseLeaseLine(ll []string, lease *LeaseRaw) error {
	if len(ll) == 0 {
		return nil
	}

	stmt := ll[0]

	switch stmt {
	case "lease":
		lease.ip = ll[1]
	case "hostname":
		lease.hostname = strings.Trim(ll[1], "\";")
	case "starts":
		date := ll[2]
		time := strings.TrimRight(ll[3], ";")
		datetime, err := parseTimeUtc(date, time)
		if err != nil {
			return err
		}
		lease.starts = datetime
	case "ends":
		date := ll[2]
		time := strings.TrimRight(ll[3], ";")
		datetime, err := parseTimeUtc(date, time)
		if err != nil {
			return err
		}
		lease.ends = datetime
	case "tstp":
		date := ll[2]
		time := strings.TrimRight(ll[3], ";")
		datetime, err := parseTimeUtc(date, time)
		if err != nil {
			return err
		}
		lease.tstp = datetime
	case "cltt":
		date := ll[2]
		time := strings.TrimRight(ll[3], ";")
		datetime, err := parseTimeUtc(date, time)
		if err != nil {
			return err
		}
		lease.cltt = datetime
	case "binding":
		lease.binding = strings.TrimRight(ll[2], ";")
	case "uid":
		lease.uid = strings.Trim(ll[1], "\";")
	case "hardware":
		lease.hardware = strings.TrimRight(ll[2], ";")
	case "set":
		// not implemented
	case "client-hostname":
		lease.clientHostname = strings.Trim(ll[1], "\";")
	case "next":
	case "rewind":
		lease.rewind = strings.TrimRight(ll[3], ";")
	default:
		return fmt.Errorf("skipping statement with no configured parser %s\n%s",
			stmt, strings.Join(ll, ","))
	}

	return nil
}

func parseTimeUtc(rawdate string, rawtime string) (time.Time, error) {
	s := fmt.Sprintf("%s %s", rawdate, rawtime)
	t, err := time.Parse("2006/01/02 15:04:05", s)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}
