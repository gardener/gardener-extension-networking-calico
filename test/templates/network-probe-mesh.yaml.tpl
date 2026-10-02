{{- /*
  network-probe-mesh.yaml.tpl — per-node HTTP server + per-node client probe mesh
  for the Calico overlay->native SeamlessOverlaySwitch integration test.

  Template values:
    .Namespace  string  namespace to deploy into (e.g. "calico-switch-test")
    .Targets    string  space-separated server pod IPs (e.g. "10.0.0.1 10.0.0.2")
                        Pass empty string on first deploy; re-render with real IPs
                        after the server DaemonSet is running.

  The client DaemonSet logs one line per probe attempt:
    <unix_ts> src=<nodeName> dst=<podIP> OK|FAIL
  Same-node flows (control) must stay OK; cross-node flows are what break during
  an overlay->native switch without the gate.
*/ -}}
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .Namespace }}
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: probe-server
  namespace: {{ .Namespace }}
  labels:
    app: probe-server
spec:
  selector:
    matchLabels:
      app: probe-server
  template:
    metadata:
      labels:
        app: probe-server
    spec:
      terminationGracePeriodSeconds: 1
      containers:
      - name: srv
        image: registry.k8s.io/e2e-test-images/agnhost:2.47
        args: ["netexec", "--http-port=8080"]
        ports:
        - containerPort: 8080
          protocol: TCP
        resources:
          requests:
            cpu: 10m
            memory: 16Mi
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: probe-client
  namespace: {{ .Namespace }}
  labels:
    app: probe-client
spec:
  selector:
    matchLabels:
      app: probe-client
  template:
    metadata:
      labels:
        app: probe-client
    spec:
      terminationGracePeriodSeconds: 1
      containers:
      - name: cli
        image: busybox:1.36
        env:
        - name: NODE
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: TARGETS
          value: "{{ .Targets }}"
        command: ["sh", "-c"]
        args:
        - |
          echo "probe-client start node=$NODE targets=${TARGETS}"
          while true; do
            ts=$(date +%s)
            for ip in $TARGETS; do
              ( if wget -T 1 -q -O /dev/null "http://${ip}:8080/hostname" 2>/dev/null; then r=OK; else r=FAIL; fi
                echo "$ts src=$NODE dst=$ip $r" ) &
            done
            wait
            sleep 0.3
          done
        resources:
          requests:
            cpu: 10m
            memory: 16Mi
