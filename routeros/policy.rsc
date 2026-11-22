# Run once. Change 192.168.88.50 to the IPv4 address of one test client.
# Add other internal prefixes to mikrot-vpn-bypass before enabling the mangle rule.
/routing/table/add name=mikrot-vpn fib
/ip/route/add dst-address=0.0.0.0/0 gateway=172.30.55.2@main routing-table=mikrot-vpn check-gateway=ping comment="mikrot-vpn gateway"
/ip/firewall/address-list/add list=mikrot-vpn-clients address=192.168.88.50 comment="mikrot-vpn test client"
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=0.0.0.0/8
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=10.0.0.0/8
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=100.64.0.0/10
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=127.0.0.0/8
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=169.254.0.0/16
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=172.16.0.0/12
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=192.168.0.0/16
/ip/firewall/address-list/add list=mikrot-vpn-bypass address=224.0.0.0/3
/ip/firewall/mangle/add chain=prerouting in-interface-list=LAN src-address-list=mikrot-vpn-clients dst-address-list=!mikrot-vpn-bypass dst-address-type=!local action=mark-routing new-routing-mark=mikrot-vpn passthrough=no disabled=yes comment="mikrot-vpn policy"
