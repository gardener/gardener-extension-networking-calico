<p>Packages:</p>
<ul>
<li>
<a href="#calico.networking.extensions.config.gardener.cloud%2fv1alpha1">calico.networking.extensions.config.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="calico.networking.extensions.config.gardener.cloud/v1alpha1">calico.networking.extensions.config.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains the Calico networking configuration API resources.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="calico.networking.extensions.config.gardener.cloud/v1alpha1.ControllerConfiguration">ControllerConfiguration
</h3>
<p>
<p>ControllerConfiguration defines the configuration for the Calico networking extension.</p>
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
<a href="https://godoc.org/k8s.io/component-base/config/v1alpha1#ClientConnectionConfiguration">
Kubernetes v1alpha1.ClientConnectionConfiguration
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ClientConnection specifies the kubeconfig file and client connection
settings for the proxy server to use when communicating with the apiserver.</p>
</td>
</tr>
<tr>
<td>
<code>healthCheckConfig</code></br>
<em>
github.com/gardener/gardener/extensions/pkg/apis/config/v1alpha1.HealthCheckConfig
</em>
</td>
<td>
<em>(Optional)</em>
<p>HealthCheckConfig is the config for the health check controller</p>
</td>
</tr>
<tr>
<td>
<code>featureGates</code></br>
<em>
map[string]bool
</em>
</td>
<td>
<em>(Optional)</em>
<p>FeatureGates is a map of feature names to bools that enable
or disable alpha/experimental features.
Default: nil</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
