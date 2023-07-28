<p>Packages:</p>
<ul>
<li>
<a href="#calico.networking.extensions.gardener.cloud%2fv1alpha1">calico.networking.extensions.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="calico.networking.extensions.gardener.cloud/v1alpha1">calico.networking.extensions.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the configuration of the Calico Network Extension.</p>
</p>
Resource Types:
<ul><li>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>
</li></ul>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig
</h3>
<p>
<p>NetworkConfig configuration for the calico networking plugin</p>
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
<code>apiVersion</code></br>
string</td>
<td>
<code>
calico.networking.extensions.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>NetworkConfig</code></td>
</tr>
<tr>
<td>
<code>backend</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.Backend">
Backend
</a>
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
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPAM">
IPAM
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPAM to use for the Calico Plugin (e.g., host-local or Calico)</p>
</td>
</tr>
<tr>
<td>
<code>ipv4</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPv4">
IPv4
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>IPv4 contains configuration for calico ipv4 specific settings</p>
</td>
</tr>
<tr>
<td>
<code>typha</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.Typha">
Typha
</a>
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
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.EbpfDataplane">
EbpfDataplane
</a>
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
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.Overlay">
Overlay
</a>
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
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.SnatToUpstreamDNS">
SnatToUpstreamDNS
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server (default: enabled)</p>
</td>
</tr>
<tr>
<td>
<code>ipip</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPv4PoolMode">
IPv4PoolMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>DEPRECATED.
IPIP is the IPIP Mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)
It was moved into the IPv4 struct, kept for backwards compatibility.
Will be removed in a future Gardener release.</p>
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
<p>DEPRECATED.
IPAutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
It was moved into the IPv4 struct, kept for backwards compatibility.
Will be removed in a future Gardener release.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.Backend">Backend
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
</p>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.CIDR">CIDR
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPAM">IPAM</a>)
</p>
<p>
</p>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.EbpfDataplane">EbpfDataplane
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
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
bool
</em>
</td>
<td>
<p>Enabled enables the eBPF dataplane mode.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.IPAM">IPAM
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
<p>IPAM defines the block that configuration for the ip assignment plugin to be used</p>
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
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.CIDR">
CIDR
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>CIDR defines the CIDR block to be used</p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.IPv4">IPv4
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
<p>IPv4 contains configuration for calico ipv4 specific settings</p>
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
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPv4Pool">
IPv4Pool
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Pool configures the type of ip pool for the tunnel interface
<a href="https://docs.projectcalico.org/v3.8/reference/node/configuration#environment-variables">https://docs.projectcalico.org/v3.8/reference/node/configuration#environment-variables</a></p>
</td>
</tr>
<tr>
<td>
<code>mode</code></br>
<em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPv4PoolMode">
IPv4PoolMode
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Mode is the mode for the IPv4 Pool (e.g. Always, Never, CrossSubnet)
ipip pools accept all pool mode values values
vxlan pools accept only Always and Never (unchecked)</p>
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
<p>AutoDetectionMethod is the method to use to autodetect the IPv4 address for this host. This is only used when the IPv4 address is being autodetected.
<a href="https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods">https://docs.projectcalico.org/v3.8/reference/node/configuration#ip-autodetection-methods</a></p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.IPv4Pool">IPv4Pool
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPv4">IPv4</a>)
</p>
<p>
</p>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.IPv4PoolMode">IPv4PoolMode
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>, 
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.IPv4">IPv4</a>)
</p>
<p>
</p>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.NetworkStatus">NetworkStatus
</h3>
<p>
<p>NetworkStatus contains information about created Network resources.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.Overlay">Overlay
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
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
bool
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
bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>CreatePodRoutes installs routes to pods on all cluster nodes.
This will only work if the cluster nodes share a single L2 network.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.SnatToUpstreamDNS">SnatToUpstreamDNS
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
<p>SnatToUpstreamDNS enables the masquerading of packets to the upstream dns server</p>
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
bool
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="calico.networking.extensions.gardener.cloud/v1alpha1.Typha">Typha
</h3>
<p>
(<em>Appears on:</em>
<a href="#calico.networking.extensions.gardener.cloud/v1alpha1.NetworkConfig">NetworkConfig</a>)
</p>
<p>
<p>Typha defines the block with configurations for calico typha</p>
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
bool
</em>
</td>
<td>
<p>Enabled is used to define whether calico-typha is required or not.
Note, typha is used to offload kubernetes API server,
thus consider not to disable it for large clusters in terms of node count.
More info can be found here <a href="https://docs.projectcalico.org/v3.9/reference/typha/">https://docs.projectcalico.org/v3.9/reference/typha/</a></p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
