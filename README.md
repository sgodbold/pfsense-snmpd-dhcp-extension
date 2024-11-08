# pfsense-snmpd-dhcp-extension
Extend pfsense SNMPD to share current DHCP leases.

```
$ snmpwalk -v 3 -l authPriv -a SHA -A PASSWORD -u USER -x AES -X PASSWORD PFSENSE_IP:16500 nsExtendOutput1
NET-SNMP-EXTEND-MIB::nsExtendOutput1Line."dhcp_leases" = STRING: {"ip":"192.168.1.1","hostname":"host-1","mac":"XX:XX:XX:XX:XX:XX"}
NET-SNMP-EXTEND-MIB::nsExtendOutputFull."dhcp_leases" = STRING: {"ip":"192.168.1.1","hostname":"host-2","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.2","hostname":"host-3","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"10.0.0.1","hostname":"host-4","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.3","hostname":"host-5","mac":"XX:XX:XX:XX:XX:XX"}
{"ip":"192.168.1.4","hostname":"host-6","mac":"XX:XX:XX:XX:XX:XX"}
NET-SNMP-EXTEND-MIB::nsExtendOutNumLines."dhcp_leases" = INTEGER: 5
NET-SNMP-EXTEND-MIB::nsExtendResult."dhcp_leases" = INTEGER: 0
```

