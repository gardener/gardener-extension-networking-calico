<p>Packages:</p>
<ul>
<li>
<a href="#calico.networking.extensions.gardener.cloud%2fv1alpha1">calico.networking.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="calico.networking.extensions.gardener.cloud/v1alpha1">calico.networking.extensions.gardener.cloud/v1alpha1</h2>
<p>

</p>

<h3 id="autoscaling">AutoScaling
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>
AutoScaling defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>mode</code></br>
<em>
<a href="#autoscalingmode">AutoscalingMode</a>
</em>
</td>
<td>
<p>Mode defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).</p>
</td>
</tr>
<tr>
<td>
<code>resources</code></br>
<em>
<a href="#resources">Resources</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Resources optionally defines the amount of resources to statically allocate for the calico components in case of<br />static resource allocation.<br />In case of vertical pod autoscaling with VPA, this field defines the minimum resources to allocate.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="autoscalingmode">AutoscalingMode
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#autoscaling">AutoScaling</a>)
</p>

<p>
AutoscalingMode is a type alias for the autoscaling mode string.
</p>


<h3 id="backend">Backend
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>


<h3 id="birdexporter">BirdExporter
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p>Enabled enables the bird metrics exporter.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="cidr">CIDR
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#ipam">IPAM</a>)
</p>

<p>

</p>


<h3 id="ebpfdataplane">EbpfDataplane
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p>Enabled enables the eBPF dataplane mode.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="ipam">IPAM
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>
IPAM defines the block that configuration for the ip assignment plugin to be used
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>type</code></br>
<em>
string
</em>
</td>
<td>
<p>Type defines the IPAM plugin type</p>
</td>
</tr>
<tr>
<td>
<code>cidr</code></br>
<em>
<a href="#cidr">CIDR</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CIDR defines the CIDR block to be used</p>
</td>
</tr>

</tbody>
</table>


<h3 id="ipv4">IPv4
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>
IPv4 contains configuration for calico ipv4 specific settings
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>pool</code></br>
<em>
<a href="#pool">Pool</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pool configures the type of ip pool for the tunnel interface<br />https://docs.projectcalico.org/v3.8/reference/node/configuration#environment-variables</p>
</td>
</tr>
<tr>
<td>
<code>mode</code></br>
<em>
<a href="#poolmode">PoolMode</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mode is the mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)<br />ipip pools accept all pool mode values values<br />vxlan pools accept only Always and Never (unchecked)</p>
</td>
</tr>
<tr>
<td>
<code>autoDetectionMethod</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.<br />https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods</p>
</td>
</tr>

</tbody>
</table>


<h3 id="ipv4pool">IPv4Pool
</h3>


<p>

</p>


<h3 id="ipv4poolmode">IPv4PoolMode
</h3>


<p>

</p>


<h3 id="ipv6">IPv6
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>
IPv6 contains configuration for calico ipv6 specific settings
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>pool</code></br>
<em>
<a href="#pool">Pool</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pool configures the type of ip pool for the tunnel interface<br />https://docs.tigera.io/calico/latest/reference/configure-calico-node#configuring-the-default-ip-pools</p>
</td>
</tr>
<tr>
<td>
<code>mode</code></br>
<em>
<a href="#poolmode">PoolMode</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mode is the mode for the IPv6 Pool (e.g. Always, Never, CrossSubnet)<br />vxlan pools accept only Always and Never (unchecked)</p>
</td>
</tr>
<tr>
<td>
<code>autoDetectionMethod</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutoDetectionMethod is the method to use to autodetect the IPv6 address for this host. This is only used when the IPv6 address is being autodetected.<br />https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods</p>
</td>
</tr>
<tr>
<td>
<code>sourceNATEnabled</code></br>
<em>
boolean
</em>
</td>
<td>
<em>(Optional)</em>
<p>SourceNATEnabled indicates whether the pod IP addresses should be masqueraded when targeting external destinations.<br />Per default, source network address translation is disabled.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="multus">Multus
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p>Enabled enables Multus CNI.</p>
</td>
</tr>
<tr>
<td>
<code>installCNIPlugins</code></br>
<em>
boolean
</em>
</td>
<td>
<em>(Optional)</em>
<p>InstallCNIPlugins enables the installation of containernetworking/plugins.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="networkconfig">NetworkConfig
</h3>


<p>
NetworkConfig configuration for the calico networking plugin
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>backend</code></br>
<em>
<a href="#backend">Backend</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Backend defines whether a backend should be used or not (e.g., bird or none)</p>
</td>
</tr>
<tr>
<td>
<code>ipam</code></br>
<em>
<a href="#ipam">IPAM</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPAM to use for the Calico Plugin (e.g., host-local or calico-ipam)</p>
</td>
</tr>
<tr>
<td>
<code>ipv4</code></br>
<em>
<a href="#ipv4">IPv4</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPv4 contains configuration for calico ipv4 specific settings</p>
</td>
</tr>
<tr>
<td>
<code>ipv6</code></br>
<em>
<a href="#ipv6">IPv6</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPv6 contains configuration for calico ipv4 specific settings</p>
</td>
</tr>
<tr>
<td>
<code>typha</code></br>
<em>
<a href="#typha">Typha</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Typha settings to use for calico-typha component</p>
</td>
</tr>
<tr>
<td>
<code>vethMTU</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>VethMTU settings used to configure calico port mtu</p>
</td>
</tr>
<tr>
<td>
<code>ebpfDataplane</code></br>
<em>
<a href="#ebpfdataplane">EbpfDataplane</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>EbpfDataplane enables the eBPF dataplane mode.</p>
</td>
</tr>
<tr>
<td>
<code>overlay</code></br>
<em>
<a href="#overlay">Overlay</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Overlay enables the network overlay</p>
</td>
</tr>
<tr>
<td>
<code>snatToUpstreamDNS</code></br>
<em>
<a href="#snattoupstreamdns">SnatToUpstreamDNS</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server (default: enabled)</p>
</td>
</tr>
<tr>
<td>
<code>autoScaling</code></br>
<em>
<a href="#autoscaling">AutoScaling</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>AutoScaling defines how the calico components are automatically scaled. It allows to use static configuration, vertical pod or cluster-proportional autoscaler (default: cluster-proportional).</p>
</td>
</tr>
<tr>
<td>
<code>vxlan</code></br>
<em>
<a href="#vxlan">VXLAN</a>
</em>
</td>
<td>
<p>VXLAN enables vxlan as overlay network</p>
</td>
</tr>
<tr>
<td>
<code>ipip</code></br>
<em>
<a href="#poolmode">PoolMode</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DEPRECATED.<br />IPIP is the IPIP Mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)<br />It was moved into the IPv4 struct, kept for backwards compatibility.<br />Will be removed in a future Gardener release.</p>
</td>
</tr>
<tr>
<td>
<code>ipAutodetectionMethod</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>DEPRECATED.<br />IPAutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.<br />It was moved into the IPv4 struct, kept for backwards compatibility.<br />Will be removed in a future Gardener release.</p>
</td>
</tr>
<tr>
<td>
<code>wireguardEncryption</code></br>
<em>
boolean
</em>
</td>
<td>
<p>WireguardEncryption is the option to enable node to node wireguard encryption</p>
</td>
</tr>
<tr>
<td>
<code>birdExporter</code></br>
<em>
<a href="#birdexporter">BirdExporter</a>
</em>
</td>
<td>
<p>BirdExporter configures the bird metrics exporter.</p>
</td>
</tr>
<tr>
<td>
<code>multus</code></br>
<em>
<a href="#multus">Multus</a>
</em>
</td>
<td>
<p>Multus configures Multus CNI.</p>
</td>
</tr>
<tr>
<td>
<code>serviceLoopPrevention</code></br>
<em>
<a href="#serviceloopprevention">ServiceLoopPrevention</a>
</em>
</td>
<td>
<p>ServiceLoopPrevention configures the Felix service loop prevention option.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="networkstatus">NetworkStatus
</h3>


<p>
NetworkStatus contains information about created Network resources.
</p>


<h3 id="overlay">Overlay
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p>Enabled enables the network overlay.</p>
</td>
</tr>
<tr>
<td>
<code>createPodRoutes</code></br>
<em>
boolean
</em>
</td>
<td>
<em>(Optional)</em>
<p>CreatePodRoutes installs routes to pods on all cluster nodes.<br />This will only work if the cluster nodes share a single L2 network.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="pool">Pool
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#ipv4">IPv4</a>, <a href="#ipv6">IPv6</a>)
</p>

<p>

</p>


<h3 id="poolmode">PoolMode
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#ipv4">IPv4</a>, <a href="#ipv6">IPv6</a>, <a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>


<h3 id="resources">Resources
</h3>


<p>
(<em>Appears on:</em><a href="#autoscaling">AutoScaling</a>)
</p>

<p>
Resources optionally defines the amount of resources to statically allocate for the calico components in case of
static resource allocation.
In case of vertical pod autoscaling with VPA, this field defines the minimum resources to allocate.
</p>


<h3 id="serviceloopprevention">ServiceLoopPrevention
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>


<h3 id="snattoupstreamdns">SnatToUpstreamDNS
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>
SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p></p>
</td>
</tr>

</tbody>
</table>


<h3 id="typha">Typha
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>
Typha defines the block with configurations for calico typha
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p>Enabled is used to define whether calico-typha is required or not.<br />Note, typha is used to offload kubernetes API server,<br />thus consider not to disable it for large clusters in terms of node count.<br />More info can be found here https://docs.projectcalico.org/v3.9/reference/typha/</p>
</td>
</tr>

</tbody>
</table>


<h3 id="vxlan">VXLAN
</h3>


<p>
(<em>Appears on:</em><a href="#networkconfig">NetworkConfig</a>)
</p>

<p>

</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>enabled</code></br>
<em>
boolean
</em>
</td>
<td>
<p>Enabled enables vxlan as overlay network.</p>
</td>
</tr>

</tbody>
</table>


