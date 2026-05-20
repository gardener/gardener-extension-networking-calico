<p>Packages:</p>
<ul>
<li>
<a href="#calico.networking.extensions.config.gardener.cloud%2fv1alpha1">calico.networking.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="calico.networking.extensions.config.gardener.cloud/v1alpha1">calico.networking.extensions.config.gardener.cloud/v1alpha1</h2>
<p>

</p>

<h3 id="controllerconfiguration">ControllerConfiguration
</h3>


<p>
ControllerConfiguration defines the configuration for the Calico networking extension.
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
<code>clientConnection</code></br>
<em>
<a href="#clientconnectionconfiguration">ClientConnectionConfiguration</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientConnection specifies the kubeconfig file and client connection<br />settings for the proxy server to use when communicating with the apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>featureGates</code></br>
<em>
object (keys:string, values:boolean)
</em>
</td>
<td>
<em>(Optional)</em>
<p>FeatureGates is a map of feature names to bools that enable<br />or disable alpha/experimental features.<br />Default: nil</p>
</td>
</tr>

</tbody>
</table>


