package kube

import (
    "encoding/json"
    "fmt"
    "sort"
    "strings"

    "k8s-s3-backup/internal/execx"
)

func Kubectl(args ...string) (string, error) { return execx.Run("kubectl", args...) }

// ResolvePodName determines a pod name based on workload kind/name or a label selector.
func ResolvePodName(ns, kind, name, selector string) (string, error) {
    if selector != "" {
        out, err := Kubectl("get", "pods", "-n", ns, "-l", selector, "-o", "jsonpath={.items[0].metadata.name}")
        return strings.TrimSpace(out), err
    }
    switch strings.ToLower(kind) {
    case "deployment", "deploy", "deployments":
        sel, err := workloadSelector(ns, "deployment", name)
        if err != nil { return "", err }
        out, err := Kubectl("get", "pods", "-n", ns, "-l", sel, "-o", "jsonpath={.items[0].metadata.name}")
        return strings.TrimSpace(out), err
    case "statefulset", "sts", "statefulsets":
        sel, err := workloadSelector(ns, "statefulset", name)
        if err != nil { return "", err }
        out, err := Kubectl("get", "pods", "-n", ns, "-l", sel, "-o", "jsonpath={.items[0].metadata.name}")
        return strings.TrimSpace(out), err
    case "pod", "pods":
        if name == "" { return "", fmt.Errorf("WORKLOAD_NAME required for pod kind") }
        if _, err := Kubectl("get", "pod", name, "-n", ns); err != nil { return "", err }
        return name, nil
    case "":
        return "", fmt.Errorf("either POD_LABEL_SELECTOR or WORKLOAD_KIND/WORKLOAD_NAME must be provided")
    default:
        return "", fmt.Errorf("unsupported WORKLOAD_KIND: %s", kind)
    }
}

// workloadSelector reads a workload and returns a comma-joined matchLabels selector.
func workloadSelector(ns, kind, name string) (string, error) {
    if name == "" { return "", fmt.Errorf("WORKLOAD_NAME is required for kind %s", kind) }
    out, err := Kubectl("get", kind, name, "-n", ns, "-o", "json")
    if err != nil { return "", err }
    var obj struct {
        Spec struct {
            Selector struct {
                MatchLabels map[string]string `json:"matchLabels"`
            } `json:"selector"`
            Template struct {
                Metadata struct{ Labels map[string]string `json:"labels"` } `json:"metadata"`
            } `json:"template"`
        } `json:"spec"`
    }
    if err := json.Unmarshal([]byte(out), &obj); err != nil { return "", err }
    labels := obj.Spec.Selector.MatchLabels
    if len(labels) == 0 { labels = obj.Spec.Template.Metadata.Labels }
    if len(labels) == 0 { return "", fmt.Errorf("no labels on %s/%s to select pods", kind, name) }
    parts := make([]string, 0, len(labels))
    for k, v := range labels { parts = append(parts, fmt.Sprintf("%s=%s", k, v)) }
    sort.Strings(parts)
    return strings.Join(parts, ","), nil
}

// DetectSingleContainer returns the container name if the pod has exactly one container.
func DetectSingleContainer(ns, pod string) (string, error) {
    out, err := Kubectl("get", "pod", pod, "-n", ns, "-o", "json")
    if err != nil { return "", err }
    var obj struct { Spec struct { Containers []struct{ Name string `json:"name"` } `json:"containers"` } `json:"spec"` }
    if err := json.Unmarshal([]byte(out), &obj); err != nil { return "", err }
    switch len(obj.Spec.Containers) {
    case 0:
        return "", fmt.Errorf("pod has no containers")
    case 1:
        return obj.Spec.Containers[0].Name, nil
    default:
        return "", fmt.Errorf("multiple containers; set CONTAINER explicitly")
    }
}
