// Package kubernetes is a Kubernetes driver for togo deploy. It renders a
// Deployment + Service (or applies provided manifests) and `kubectl apply`s them
// to the cluster in KUBECONFIG. Select with DEPLOY_PROVIDER=kubernetes.
package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/togo-framework/deploy"
	"github.com/togo-framework/togo"
)

func init() {
	deploy.RegisterDriver("kubernetes", func(k *togo.Kernel) (deploy.Deployer, error) {
		ns := os.Getenv("KUBE_NAMESPACE")
		if ns == "" {
			ns = "default"
		}
		return &driver{namespace: ns}, nil
	})
}

type driver struct{ namespace string }

func kubectl(ctx context.Context, stdin string, args ...string) (string, error) {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return "", fmt.Errorf("deploy-kubernetes: kubectl not found on PATH")
	}
	cmd := exec.CommandContext(ctx, "kubectl", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	return out.String(), cmd.Run()
}

func (d *driver) manifest(spec deploy.Spec) string {
	img := spec.Image
	if img == "" {
		img = spec.App + ":latest"
	}
	var env bytes.Buffer
	for k, v := range spec.Env {
		fmt.Fprintf(&env, "        - name: %s\n          value: %q\n", k, v)
	}
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata: { name: %[1]s, namespace: %[2]s }
spec:
  replicas: 1
  selector: { matchLabels: { app: %[1]s } }
  template:
    metadata: { labels: { app: %[1]s } }
    spec:
      containers:
      - name: %[1]s
        image: %[3]s
        ports: [{ containerPort: 8080 }]
        env:
%[4]s---
apiVersion: v1
kind: Service
metadata: { name: %[1]s, namespace: %[2]s }
spec:
  selector: { app: %[1]s }
  ports: [{ port: 80, targetPort: 8080 }]
`, spec.App, d.namespace, img, env.String())
}

func (d *driver) Provision(ctx context.Context, spec deploy.Spec) (*deploy.Result, error) {
	if out, err := kubectl(ctx, "", "create", "namespace", d.namespace, "--dry-run=client", "-o", "yaml"); err != nil {
		return nil, fmt.Errorf("kubernetes: %w\n%s", err, out)
	}
	_, _ = kubectl(ctx, "", "create", "namespace", d.namespace)
	return &deploy.Result{Message: "namespace " + d.namespace + " ready"}, nil
}

func (d *driver) Deploy(ctx context.Context, spec deploy.Spec) (*deploy.Result, error) {
	m := d.manifest(spec)
	if p, ok := spec.Options["manifest"].(string); ok && p != "" {
		m = p
	}
	if out, err := kubectl(ctx, m, "apply", "-f", "-"); err != nil {
		return nil, fmt.Errorf("kubectl apply: %w\n%s", err, out)
	}
	return &deploy.Result{Message: "applied " + spec.App + " to ns/" + d.namespace}, nil
}

func (d *driver) Destroy(ctx context.Context, spec deploy.Spec) error {
	_, err := kubectl(ctx, "", "delete", "deployment,service", spec.App, "-n", d.namespace, "--ignore-not-found")
	return err
}

func (d *driver) Status(ctx context.Context, spec deploy.Spec) (*deploy.Status, error) {
	out, err := kubectl(ctx, "", "get", "deployment", spec.App, "-n", d.namespace, "-o", "jsonpath={.status.availableReplicas}")
	healthy := err == nil && strings.TrimSpace(out) != "" && strings.TrimSpace(out) != "0"
	return &deploy.Status{Healthy: healthy, Detail: "availableReplicas=" + strings.TrimSpace(out)}, nil
}
