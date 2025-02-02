# pfsense-snmpd-dhcp-extension
Extend pfsense SNMPD to share current DHCP leases. I created this to be consumed by my CoredDNS plugin to serve SNMP data as DNS A records https://github.com/sgodbold/coredns-snmp-dhcp-plugin

```
$ snmpwalk -On -v 3 -l authPriv -a SHA -A PASSWORD -u USER -x AES -X PASSWORD PFSENSE_IP:16500 nsExtendOutput1
NET-SNMP-EXTEND-MIB::nsExtendOutput1Line."dhcp_leases" = STRING: {"ip":"192.168.1.1","hostname":"host-1","mac":"XX:XX:XX:XX:XX:XX"}
NET-SNMP-EXTEND-MIB::nsExtendOutputFull."dhcp_leases" = STRING: {"ip":"192.168.1.1","hostname":"host-2","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.2","hostname":"host-3","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"10.0.0.1","hostname":"host-4","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.3","hostname":"host-5","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.4","hostname":"host-6","mac":"XX:XX:XX:XX:XX:XX"}
NET-SNMP-EXTEND-MIB::nsExtendOutNumLines."dhcp_leases" = INTEGER: 5
NET-SNMP-EXTEND-MIB::nsExtendResult."dhcp_leases" = INTEGER: 0

ALT OUTPUT (i think if you don't have the pfsense MIBS loaded
.1.3.6.1.4.1.8072.1.3.2.3.1.4.6.115.116.101.118.101.110 = STRING: {"ip":"192.168.1.1","hostname":"host-1","mac":"XX:XX:XX:XX:XX:XX"}
.1.3.6.1.4.1.8072.1.3.2.3.1.4.6.115.116.101.118.101.110 = STRING: {"ip":"192.168.1.1","hostname":"host-2","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.2","hostname":"host-3","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"10.0.0.1","hostname":"host-4","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.3","hostname":"host-5","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.4","hostname":"host-6","mac":"XX:XX:XX:XX:XX:XX"}
.1.3.6.1.4.1.8072.1.3.2.3.1.4.6.115.116.101.118.101.110 = INTEGER: 5
.1.3.6.1.4.1.8072.1.3.2.3.1.4.6.115.116.101.118.101.110 = INTEGER: 0
```

