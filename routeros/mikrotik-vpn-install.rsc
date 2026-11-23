# Run in an interactive RouterOS terminal with /import file-name=mikrotik-vpn-install.rsc.
{
    :global mikrotikVpnReleaseBase
    :local releaseBase [:tostr $mikrotikVpnReleaseBase]
    :set mikrotikVpnReleaseBase
    :if ($releaseBase != "") do={
        :if ([:pick $releaseBase 0 8] != "https://") do={ :error "The release base must be an HTTPS URL." }
        :if ([:pick $releaseBase ([:len $releaseBase] - 1) [:len $releaseBase]] = "/") do={
            :set releaseBase [:pick $releaseBase 0 ([:len $releaseBase] - 1)]
        }
    }
    :local serviceName "mikrot-vpn"
    :local vethName "veth-mikrot"
    :local dataList "mikrot-data"
    :local envListName "mikrot-env"
    :local clientList "mikrot-vpn-clients"
    :local bypassList "mikrot-vpn-bypass"
    :local panelChain "mikrot-vpn-access"
    :local connectionTag "mikrotik-vpn-route"
    :local ownerTag "MikroTik VPN installer"

    :onerror packageError in={ /container/print count-only } do={
        :error "Install the RouterOS container package matching your RouterOS version, then reboot."
    }
    :if ([/system/device-mode/get container] != true) do={
        :error "Enable containers with /system/device-mode/update container=yes and confirm it on the router."
    }
    :if ([:len [/container/find where name=$serviceName]] > 0) do={
        :error "MikroTik VPN already exists. This installer is for a fresh installation."
    }
    :if (([:len [/interface/find where name=$vethName]] > 0) || ([:len [/routing/table/find where name=$serviceName]] > 0) || ([:len [/container/mounts/find where list=$dataList]] > 0) || ([:len [/container/envs/find where list=$envListName]] > 0) || ([:len [/ip/firewall/address-list/find where list=$clientList]] > 0) || ([:len [/ip/firewall/address-list/find where list=$bypassList]] > 0)) do={
        :error "MikroTik VPN resource names are already in use. Inspect the existing installation before retrying."
    }
    :if (([:len [/ip/firewall/filter/find where comment=$ownerTag or chain=$panelChain]] > 0) || ([:len [/ip/firewall/nat/find where comment=$ownerTag]] > 0) || ([:len [/ip/firewall/mangle/find where comment=$ownerTag]] > 0)) do={
        :error "Rules from a previous installation already exist."
    }

    :local routerArch [/system/resource/get architecture-name]
    :local archiveName ""
    :if ($routerArch = "arm") do={
        :set archiveName "mikrot-armv7.tar"
        :if ([/system/resource/get board-name] = "hEX") do={ :set archiveName "mikrot-armv5.tar" }
    }
    :if ($routerArch = "arm64") do={ :set archiveName "mikrot-arm64.tar" }
    :if ($routerArch = "x86_64") do={ :set archiveName "mikrot-amd64.tar" }
    :if ($archiveName = "") do={ :error ("Unsupported architecture: " . $routerArch) }

    :put ("RouterOS architecture: " . $routerArch)
    :put "Mounted writable disks with space for the container:"
    :local disks [:toarray ""]
    :local bestDisk ""
    :local bestFree 0
    :local bestHasArchive false
    :local reserveBytes 100663296
    :foreach device in=[/disk/find where mounted and !mount-read-only] do={
        :local mountName [/disk/get $device mount-point]
        :local fsType [/disk/get $device fs]
        :local freeBytes [/disk/get $device free]
        :if (($mountName != "") && ($mountName != "flash") && ($freeBytes >= $reserveBytes) && (($fsType = "ext4") || ($fsType = "xfs") || ($fsType = "btrfs"))) do={
            :set ($disks->$mountName) $freeBytes
            :put ("  " . $mountName . " (" . $fsType . ", " . ($freeBytes / 1048576) . " MiB free)")
            :local hasArchive (($releaseBase = "") && ([:len [/file/find where name=($mountName . "/" . $archiveName)]] > 0))
            :if ((($hasArchive = true) && ($bestHasArchive = false)) || (($hasArchive = $bestHasArchive) && ($freeBytes > $bestFree))) do={
                :set bestDisk $mountName
                :set bestFree $freeBytes
                :set bestHasArchive $hasArchive
            }
        }
    }
    :if ($bestDisk = "") do={ :error "No suitable mounted disk. Use a writable ext4, XFS or Btrfs disk with at least 96 MiB free." }
    :put ("Press Enter to install on " . $bestDisk . ".")
    :local selectedDisk [/terminal/ask value-name="Install on disk: "]
    :if ($selectedDisk = "") do={ :set selectedDisk $bestDisk }
    :if ([:typeof ($disks->$selectedDisk)] != "num") do={ :error "Enter a mount point from the list above." }

    :local defaultSource ($selectedDisk . "/" . $archiveName)
    :if ($releaseBase != "") do={ :set defaultSource ($releaseBase . "/" . $archiveName) }
    :put ("Enter a file path or HTTPS URL. Default: " . $defaultSource)
    :local imageSource [/terminal/ask value-name="Container archive path: "]
    :if ($imageSource = "") do={ :set imageSource $defaultSource }
    :local fetchImage ([:pick $imageSource 0 8] = "https://")
    :if (!$fetchImage && ([:len [/file/find where name=$imageSource]] = 0)) do={
        :error ("Archive not found: " . $imageSource)
    }

    :local defaultSubnet "172.30.55.0/30"
    :local subnetText [/terminal/ask value-name=("Container IPv4 subnet [" . $defaultSubnet . "]: ")]
    :if ($subnetText = "") do={ :set subnetText $defaultSubnet }
    :local subnetSlash [:find $subnetText "/"]
    :if ([:typeof $subnetSlash] != "num") do={ :error "Enter an IPv4 subnet in CIDR notation." }
    :local subnetBits [:tonum [:pick $subnetText ($subnetSlash + 1) [:len $subnetText]]]
    :if (([:typeof $subnetBits] != "num") || ($subnetBits < 1) || ($subnetBits > 30)) do={
        :error "Enter a dedicated IPv4 subnet with at least two usable addresses."
    }
    :local networkIp [:toip [:pick $subnetText 0 $subnetSlash]]
    :local networkMask (255.255.255.255 << (32 - $subnetBits))
    :if ([:typeof $networkIp] != "ip") do={ :error "Invalid IPv4 subnet." }
    :if (($networkIp & $networkMask) != $networkIp) do={ :error "Enter the network address of the subnet." }
    :if (($networkIp in 0.0.0.0/8) || ($networkIp in 127.0.0.0/8) || ($networkIp in 224.0.0.0/3) || ($networkIp in 198.18.0.0/15) || ((127.0.0.0 & $networkMask) = $networkIp) || ((224.0.0.0 & $networkMask) = $networkIp) || ((198.18.0.0 & $networkMask) = $networkIp)) do={
        :error "Choose an IPv4 subnet outside reserved and TUN ranges."
    }
    :local subnet ($networkIp . "/" . $subnetBits)
    :local routerIp ($networkIp | 0.0.0.1)
    :local containerIp ($networkIp | 0.0.0.2)
    :foreach routeId in=[/ip/route/find where routing-table=main and dst-address!=0.0.0.0/0] do={
        :local routePrefix [/ip/route/get $routeId dst-address]
        :local routeText [:tostr $routePrefix]
        :local routeIp [:toip [:pick $routeText 0 [:find $routeText "/"]]]
        :if (($networkIp in $routePrefix) || (($routeIp & $networkMask) = $networkIp)) do={
            :error ("Container subnet overlaps RouterOS route " . $routeText)
        }
    }

    :local lanList [/terminal/ask value-name="Trusted LAN interface list [LAN]: "]
    :if ($lanList = "") do={ :set lanList "LAN" }
    :local wanList [/terminal/ask value-name="Internet/WAN interface list [WAN]: "]
    :if ($wanList = "") do={ :set wanList "WAN" }
    :local builtinLists "^(all|none|dynamic|static)\$"
    :if (($lanList = $wanList) || ($lanList ~ $builtinLists) || ($wanList ~ $builtinLists) || ([:len [/interface/list/find where name=$lanList]] = 0) || ([:len [/interface/list/find where name=$wanList]] = 0)) do={
        :error "Choose distinct existing interface lists for the trusted LAN and WAN."
    }

    :put "Press Enter to install the panel without enabling client routing."
    :local clientInput [/terminal/ask value-name="Client IPv4 address or CIDR: "]
    :local clientSlash [:find $clientInput "/"]
    :if ($clientInput != "") do={
        :if ([:typeof $clientSlash] = "num") do={
            :local clientBits [:tonum [:pick $clientInput ($clientSlash + 1) [:len $clientInput]]]
            :if (([:typeof $clientBits] != "num") || ($clientBits < 1) || ($clientBits > 32) || ([:typeof [:toip [:pick $clientInput 0 $clientSlash]]] != "ip")) do={
                :error "Enter a valid client IPv4 subnet."
            }
        } else={
            :if ([:typeof [:toip $clientInput]] != "ip") do={ :error "Enter a valid client IPv4 address." }
        }
    }
    :put "Enter a password of 12 or more characters, or press Enter to generate one. Typed characters are visible."
    :local panelPassword [/terminal/ask value-name="Panel password: "]
    :local generatedPassword ($panelPassword = "")
    :if ($generatedPassword) do={ :set panelPassword [:rndstr length=24] }
    :if (([:len $panelPassword] < 12) || ([:len $panelPassword] > 128)) do={ :error "Panel password must be between 12 and 128 characters." }

    :local stateDir ($selectedDisk . "/mikrot-data")
    :local rootDir ($selectedDisk . "/mikrot-root")
    :if ([:len [/file/find where name=$rootDir]] > 0) do={ :error ("Existing container root directory: " . $rootDir) }
    :if ($fetchImage) do={
        :local downloadPath ($selectedDisk . "/mikrot-download-" . [:rndstr length=8] . ".tar")
        :if ([:len [/file/find where name=$downloadPath]] > 0) do={ :error "Temporary archive path exists. Retry the installation." }
        :onerror downloadError in={
            /tool/fetch url=$imageSource dst-path=$downloadPath check-certificate=yes http-max-redirect-count=5
        } do={
            :do { /file/remove [/file/find where name=$downloadPath] } on-error={}
            :error ("Archive download failed: " . $downloadError)
        }
        :set imageSource $downloadPath
    }
    :local archiveId [/file/find where name=$imageSource]
    :if ([:len $archiveId] != 1) do={ :error "Container archive is not available." }
    :local archiveBytes [/file/get $archiveId size]
    :local selectedDiskId [/disk/find where mount-point=$selectedDisk and mounted and !mount-read-only]
    :if ([:len $selectedDiskId] != 1) do={ :error "Target disk is no longer mounted." }
    :if (($archiveBytes <= 0) || ([/disk/get $selectedDiskId free] < ($archiveBytes + 16777216))) do={
        :error "The archive is empty or there is insufficient space to extract it."
    }
    :if ([:len [/file/find where name=$stateDir]] = 0) do={ /file/add name=$stateDir type=directory }

    :put ("Installing on " . $selectedDisk . ", container subnet " . $subnet)
    :onerror installError in={
        /interface/veth/add name=$vethName address=($containerIp . "/" . $subnetBits) gateway=$routerIp comment=$ownerTag
        /ip/address/add address=($routerIp . "/" . $subnetBits) interface=$vethName comment=$ownerTag
        /container/mounts/add list=$dataList src=$stateDir dst=/data
        /container/envs/add list=$envListName key=MIKROTIK_PASSWORD value=$panelPassword

        /ip/firewall/filter/add chain=$panelChain in-interface-list=$lanList protocol=tcp dst-port=53,8080 action=accept comment=$ownerTag
        /ip/firewall/filter/add chain=$panelChain in-interface-list=$lanList protocol=udp dst-port=53 action=accept comment=$ownerTag
        /ip/firewall/filter/add chain=$panelChain protocol=tcp dst-port=53,8080 action=drop comment=$ownerTag
        /ip/firewall/filter/add chain=$panelChain protocol=udp dst-port=53 action=drop comment=$ownerTag
        /ip/firewall/filter/add chain=$panelChain action=return comment=$ownerTag
        :local firstFilter [:pick [/ip/firewall/filter/find where chain=forward and !dynamic] 0]
        :if ([:typeof $firstFilter] = "id") do={
            /ip/firewall/filter/add chain=forward out-interface=$vethName dst-address=$containerIp action=jump jump-target=$panelChain place-before=$firstFilter comment=$ownerTag
        } else={
            /ip/firewall/filter/add chain=forward out-interface=$vethName dst-address=$containerIp action=jump jump-target=$panelChain comment=$ownerTag
        }
        /ip/firewall/nat/add chain=srcnat src-address=$containerIp out-interface-list=$wanList action=masquerade comment=$ownerTag
        /routing/table/add name=$serviceName fib
        /ip/route/add dst-address=0.0.0.0/0 gateway=($containerIp . "@main") routing-table=$serviceName check-gateway=ping comment=$ownerTag
        :local bypassRanges {"0.0.0.0/8";"10.0.0.0/8";"100.64.0.0/10";"127.0.0.0/8";"169.254.0.0/16";"172.16.0.0/12";"192.168.0.0/16";"224.0.0.0/3"}
        :foreach localRange in=($bypassRanges, $subnet) do={
            /ip/firewall/address-list/add list=$bypassList address=$localRange comment=$ownerTag
        }
        :if ($clientInput != "") do={ /ip/firewall/address-list/add list=$clientList address=$clientInput comment=$ownerTag }

        :local connectionRule [/ip/firewall/mangle/add chain=prerouting in-interface-list=$lanList src-address-list=$clientList dst-address-list=("!" . $bypassList) dst-address-type=!local connection-state=new connection-mark=no-mark action=mark-connection new-connection-mark=$connectionTag passthrough=yes disabled=yes comment=$ownerTag]
        :local routingRule [/ip/firewall/mangle/add chain=prerouting in-interface-list=$lanList src-address-list=$clientList dst-address-list=("!" . $bypassList) dst-address-type=!local action=mark-routing new-routing-mark=$serviceName passthrough=no disabled=yes comment=$ownerTag]
        :local firstFastTrack [:pick [/ip/firewall/filter/find where chain=forward and action=fasttrack-connection and !disabled] 0]
        :if ([:typeof $firstFastTrack] = "id") do={
            /ip/firewall/filter/add chain=forward connection-state=established,related connection-mark=$connectionTag action=accept place-before=$firstFastTrack comment=$ownerTag
        }

        :local containerDns [/ip/dns/get servers]
        :if ([:len $containerDns] = 0) do={ :set containerDns [/ip/dns/get dynamic-servers] }
        :if ([:len $containerDns] = 0) do={ :set containerDns {"1.1.1.1";"9.9.9.9"} }
        /container/add name=$serviceName file=$imageSource interface=$vethName root-dir=$rootDir envlist=$envListName mountlists=$dataList user=0:0 dns=$containerDns logging=yes start-on-boot=no restart-interval=10s
        :local containerId [/container/find where name=$serviceName]
        :local ready false
        :local attempts 0
        :while ((!$ready) && ($attempts < 180)) do={
            :set ready ([/container/get $containerId stopped] = true)
            :if (!$ready) do={ :set attempts ($attempts + 1); :delay 1s }
        }
        :if (!$ready) do={ :error "Container extraction timed out. Check the RouterOS container log." }
        /container/set $containerId start-on-boot=yes
        /container/start $containerId
        :set ready false
        :set attempts 0
        :while ((!$ready) && ($attempts < 25)) do={
            :set ready (([/container/get $containerId running] = true) && ([/ping address=$containerIp count=1 interval=500ms] > 0))
            :if (!$ready) do={ :set attempts ($attempts + 1); :delay 1s }
        }
        :if (!$ready) do={ :error "Container did not start. Check the RouterOS log and the image architecture." }
        :if ($clientInput != "") do={
            /ip/firewall/mangle/enable $connectionRule
            /ip/firewall/mangle/enable $routingRule
        }
    } do={
        :put ("Installation failed: " . $installError)
        :local failedContainer [/container/find where name=$serviceName]
        :if ([:len $failedContainer] > 0) do={
            :do { /container/stop $failedContainer } on-error={}
            :do { /container/remove $failedContainer } on-error={}
        }
        :do { /ip/firewall/mangle/remove [/ip/firewall/mangle/find where comment=$ownerTag] } on-error={}
        :do { /ip/firewall/filter/remove [/ip/firewall/filter/find where comment=$ownerTag] } on-error={}
        :do { /ip/firewall/nat/remove [/ip/firewall/nat/find where comment=$ownerTag] } on-error={}
        :do { /ip/firewall/address-list/remove [/ip/firewall/address-list/find where comment=$ownerTag] } on-error={}
        :do { /ip/route/remove [/ip/route/find where comment=$ownerTag] } on-error={}
        :do { /routing/table/remove [/routing/table/find where name=$serviceName] } on-error={}
        :do { /container/envs/remove [/container/envs/find where list=$envListName] } on-error={}
        :do { /container/mounts/remove [/container/mounts/find where list=$dataList] } on-error={}
        :do { /ip/address/remove [/ip/address/find where comment=$ownerTag] } on-error={}
        :do { /interface/veth/remove [/interface/veth/find where name=$vethName] } on-error={}
        :error $installError
    }
    :put ("Panel: http://" . $containerIp . ":8080 (user: admin)")
    :if ($generatedPassword) do={ :put ("Generated panel password: " . $panelPassword) }
    :if ($clientInput = "") do={ :put "Client routing is disabled. The web panel is ready." }
}
