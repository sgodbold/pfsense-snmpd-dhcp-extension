package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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
	Hostname string `json:"hostname"`
	Mac      string `json:"mac"`
}

const (
	LeaseStartWord = "lease"
	LeaseEndWord   = "}"
)

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(w io.Writer, args []string) error {
	if len(args[1:]) < 1 {
		return errors.New("incorrect number of arguments")
	}

	path := args[1]

	file, err := os.Open(path)
	if err != nil {
		return errors.New(
			fmt.Sprintf("could not open file %s: %s", path, err))
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	writer := json.NewEncoder(w)

	leases, err := parseLeaseFile(reader)
	if err != nil {
		return err
	}

	printLeases(leases, writer)

	return nil
}

func parseLeaseFile(r *bufio.Reader) (map[string]*Lease, error) {
	if err := parseHeader(r); err != nil {
		return nil, errors.New(
			fmt.Sprintf("could not consume header of leases file: %s", err))
	}

	leases := make(map[string]*Lease)

	for {
		raw, err := parseLease(r)
		if err != nil {
			return nil, err
		}
		if raw == nil {
			break
		}

		if leaseFilter(raw) == false {
			lease := leaseRawMap(raw)
			leases[lease.Hostname] = lease
		}
	}

	return leases, nil
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

func leaseRawMap(l *LeaseRaw) *Lease {
	pl := newPrintableLease()

	pl.Ip = l.ip
	pl.Mac = l.hardware

	if len(l.hostname) > 0 {
		pl.Hostname = strings.ToLower(l.hostname)
	}
	if len(l.clientHostname) > 0 {
		pl.Hostname = strings.ToLower(l.clientHostname)
	}

	return pl
}

func newLeaseRaw() *LeaseRaw {
	return &LeaseRaw{}
}

func newPrintableLease() *Lease {
	return &Lease{}
}

func printLeases(leases map[string]*Lease, w *json.Encoder) {
	for _, l := range leases {
		w.Encode(l)
	}
}

func parseHeader(r *bufio.Reader) error {
	for {
		if eq, _ := peekEq(LeaseStartWord, r); eq {
			return nil
		}

		if _, err := r.ReadString('\n'); err != nil {
			return errors.New("failed to find end of header")
		}
	}
}

func parseLease(r *bufio.Reader) (*LeaseRaw, error) {
	l := newLeaseRaw()

	for {
		eq, err := peekEq(LeaseEndWord, r)

		if err != nil {
			// can't peek because EOF
			return nil, nil
		}

		if eq {
			if _, err := r.ReadString('\n'); err != nil {
				return nil, err
			}
			return l, nil
		}

		word, err := readWord(r)
		if err != nil {
			return nil, err
		}

		switch word {
		case "lease":
			ip, err := parseIp(r)
			if err != nil {
				return nil, err
			}
			l.ip = ip
		case "hostname":
			hostname, err := parseHostname(r)
			if err != nil {
				return nil, err
			}
			l.hostname = hostname
		case "starts":
			starts, err := parseTimeUtc(r)
			if err != nil {
				return nil, err
			}
			l.starts = starts
		case "ends":
			ends, err := parseTimeUtc(r)
			if err != nil {
				return nil, err
			}
			l.ends = ends
		case "tstp":
			tstp, err := parseTimeUtc(r)
			if err != nil {
				return nil, err
			}
			l.tstp = tstp
		case "cltt":
			cltt, err := parseTimeUtc(r)
			if err != nil {
				return nil, err
			}
			l.cltt = cltt
		case "binding":
			binding, err := parseBinding(r)
			if err != nil {
				return nil, err
			}
			l.binding = binding
		case "uid":
			uid, err := parseUid(r)
			if err != nil {
				return nil, err
			}
			l.uid = uid
		case "hardware":
			mac, err := parseHardware(r)
			if err != nil {
				return nil, err
			}
			l.hardware = mac
		case "set":
			// not implemented
		case "client-hostname":
			ch, err := parseClientHostname(r)
			if err != nil {
				return nil, err
			}
			l.clientHostname = ch
		case "next":
			next, err := parseNext(r)
			if err != nil {
				return nil, err
			}
			l.next = next
		case "rewind":
			rewind, err := parseRewind(r)
			if err != nil {
				return nil, err
			}
			l.rewind = rewind
		default:
			line, _ := r.ReadString('\n')
			return nil, errors.New(
				fmt.Sprintf("no parser for word %s\n%s", word, line))
		}

		_, err = r.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}

	return l, nil
}

func parseIp(r *bufio.Reader) (string, error) {
	ip, err := readWord(r)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("failed to read ip: %s", err))
	}
	return ip, nil
}

func parseHostname(r *bufio.Reader) (string, error) {
	hostname, err := readWord(r)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("failed to read hostname: %s", err))
	}
	return hostname, nil
}

func parseTimeUtc(r *bufio.Reader) (time.Time, error) {
	// parse day of week; unused
	if _, err := readWord(r); err != nil {
		return time.Now(), err
	}
	line, err := r.ReadString(';')
	if err != nil {
		return time.Now(), err
	}
	trimmed := strings.Trim(line, ";")
	t, err := time.Parse("2006/01/02 15:04:05", trimmed)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func parseBinding(r *bufio.Reader) (string, error) {
	// reads 'state'; unused
	if _, err := readWord(r); err != nil {
		return "", err
	}
	word, err := r.ReadString(';')
	if err != nil {
		return "", err
	}
	trimmed := strings.Trim(word, ";")
	return trimmed, nil
}

func parseUid(r *bufio.Reader) (string, error) {
	word, err := r.ReadString(';')
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("failed to read UID: %s", err))
	}
	trimmed := strings.Trim(word, "\";")
	return trimmed, nil
}

func parseHardware(r *bufio.Reader) (string, error) {
	// read hardware-type; unused
	if _, err := readWord(r); err != nil {
		return "", err
	}
	word, err := r.ReadString(';')
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("failed to read hardware: %s", err))
	}
	trimmed := strings.Trim(word, ";")
	return trimmed, nil
}

func parseClientHostname(r *bufio.Reader) (string, error) {
	word, err := r.ReadString(';')
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("failed to read client-hostname: %s", err))
	}
	trimmed := strings.Trim(word, "\";")
	return trimmed, nil
}

func parseNext(r *bufio.Reader) (string, error) {
	// read binding; unused
	if _, err := readWord(r); err != nil {
		return "", err
	}
	// read state; unused
	if _, err := readWord(r); err != nil {
		return "", err
	}
	word, err := r.ReadString(';')
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to read next: %s", err))
	}
	trimmed := strings.Trim(word, ";")
	return trimmed, nil
}

func parseRewind(r *bufio.Reader) (string, error) {
	// read binding; unused
	if _, err := readWord(r); err != nil {
		return "", err
	}
	// read state; unused
	if _, err := readWord(r); err != nil {
		return "", err
	}
	word, err := r.ReadString(';')
	if err != nil {
		return "", errors.New(fmt.Sprintf("failed to read next: %s", err))
	}
	trimmed := strings.Trim(word, ";")
	return trimmed, nil
}

func peekEq(s string, r *bufio.Reader) (bool, error) {
	sbar := []byte(s)

	bar, err := r.Peek(len(sbar))
	if err != nil {
		return false, err
	}

	return bytes.Equal(sbar, bar), nil
}

func readWord(r *bufio.Reader) (string, error) {
	for {
		word, err := r.ReadString(' ')
		if err != nil {
			return "", errors.New(fmt.Sprintf("failed to read word: %s", err))
		}

		trimmed := strings.Trim(word, " ")
		if len(trimmed) > 0 {
		    return trimmed, nil
		}
	}

	return "", errors.New("failed to find any word")
}
