# Frequently Asked Questions

Depending on the context you're using hcdiag in, you may have some questions about the tool. Here are questions we see frequently.

### Why is hcdiag checking for all HashiCorp products when I only want one?

We heard feedback from `hcdiag` users that the default behavior, with no flags provided, could be confusing. In that case,
it only gathered host details from the machine where it ran. This could cause users to think that data had been
collected about their HashiCorp products when it actually hadn't.

Beginning in version `0.4.0`, the default behavior - when no product flags are provided (such as `-consul`, `-vault`, etc.) -
is to detect which product CLIs are installed on the system. Any that are found are added, automatically, to the list of
products for which to pull diagnostics.

In the event that you run `hcdiag` with no product flags on a system where the CLI for multiple HashiCorp products are installed,
but any of them are not configured, this can lead to failures. For example, if you had both Terraform and Vault CLIs installed,
but only Vault was configured, then `hcdiag` might fail on Terraform Enterprise checks. A log message is displayed when auto-detection
is used, and it will indicate which products were found on your system. If you find that you wish to limit which products
are executed, please use the appropriate product flag. In the previous example, you would want to run `hcdiag -vault` instead
of just `hcdiag` because you have both CLIs on your system, but Terraform is not actually configured for use.

### Can I use hcdiag as a real-time metrics-gathering or sampling tool?

hcdiag is designed for use cases like troubleshooting, which require historical data. As a result, hcdiag retrieves data for a particular range of time in the past. It is not designed for real-time or sampling use cases where it continues to run perpetually every x minutes, reporting on changes in environment data.

### How do I use hcdiag with Kubernetes?

Although Kubernetes is a complex topic with many configuration options, the key to remember is that hcdiag must be able
to communicate with your pod(s) via a network connection in order to get diagnostic details from the products running in
those pods. If the management interface is not already exposed externally from k8s, you may consider setting up a port-forward
when collecting diagnostics. The command would be similar to `kubectl -n consul port-forward consul-server-0 8500:8500`; 
in this example, we are in the namespace `consul` (noted by `-n consul`), and we are forwarding the external port `8500`
(the port before the `:`) to port `8500` (the port after the `:`) on the pod `consul-server-0`.

If you would like to experiment with a setup like this, assuming you have both minikube and helm installed on your machine,
you could use the following as a reference:

```shell
minikube start
helm repo add hashicorp https://helm.releases.hashicorp.com
helm search repo hashicorp/consul
helm install consul hashicorp/consul --set global.name=consul --create-namespace -n consul
kubectl -n consul port-forward consul-server-0 8500:8500

# Now, run hcdiag
hcdiag -consul
```

### How can I split large bundle files into smaller ones?
If you're diagnosing issues on large clusters, hcdiag's output bundles may contain large, but necessary, amounts of data (logs, debug output, etc.).

If you need to transfer these large files, for example between a customer and a support team, we recommend sending the file in a way that's secure and encrypted end-to-end. A secure file-sharing platform such as [SendSafely](https://www.sendsafely.com/) fits this requirement.

If you are forced to use less secure and more limited methods of transfer, such as email, you can split the bundle file with a tool such as `split`, which is built into (or available on) most Unix-like systems, including Linux and Mac OS.

```
split -b 10M hcdiag-2022-09-01T133045Z.tar.gz hcsplit
```
Let's break down the arguments, piece by piece:

* `-b 10M`: split into 10-megabyte files
* `hcdiag-2022-09-01T133045Z.tar.gz`: name of the input file you'd like to split
* `hcsplit`: filename prefix that your split files should have (this is optional; the default is `x`)

When you run this command on a large bundle, you'll see something like this:

```
$ root@3891043f2342:/tmp/vault-hcdiag# ls -alh
...
-rw-r--r-- 1 root root 20M  Sep  1 13:30 hcdiag-2022-09-01T133045Z.tar.gz
-rw-r--r-- 1 root root 10M  Sep  1 13:38 hcsplitaa
-rw-r--r-- 1 root root 10M  Sep  1 13:38 hcsplitab
-rw-r--r-- 1 root root  37K Sep  1 13:38 hcsplitac
```

Your original bundle file remains untouched, and has been copied/split into these three chunks.

After the split file has been transferred to its destination, it can be reconstituted from these smaller parts with a single command: just concatenate them together using `cat`.

```
cat hcsplit* > reconstituted_bundle.tar.gz
```

You can work with this file normally, now; the file content in `reconstituted_bundle.tar.gz` is identical to that of `hcdiag-2022-09-01T133045Z.tar.gz`:

```
root@3891043f2342:/tmp/vault-hcdiag# md5sum *.tar.gz
017bed533ecbf4745edb29b832d755c8  hcdiag-2022-09-01T133045Z.tar.gz
017bed533ecbf4745edb29b832d755c8  reconstituted_bundle.tar.gz
```
