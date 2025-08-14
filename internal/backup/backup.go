package backup

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "os/exec"
    "strings"
    "time"

    "k8s-s3-backup/internal/config"
    "k8s-s3-backup/internal/kube"
    s3util "k8s-s3-backup/internal/s3"
)

func firstNonEmpty(vals ...string) string {
    for _, v := range vals {
        if strings.TrimSpace(v) != "" { return v }
    }
    return ""
}

func sanitizeName(s string) string {
    s = strings.TrimSpace(strings.ToLower(s))
    if s == "" { return "backup" }
    var b strings.Builder
    for _, r := range s {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
            b.WriteRune(r)
        } else { b.WriteByte('-') }
    }
    return b.String()
}

func Run(cfg config.Config) error {
    // 1) Resolve target pod
    pod, err := kube.ResolvePodName(cfg.Namespace, cfg.WorkloadKind, cfg.WorkloadName, cfg.LabelSelector)
    if err != nil || pod == "" { return fmt.Errorf("resolve pod: %w", err) }

    // 2) Resolve container (auto if single container)
    container := cfg.Container
    if strings.TrimSpace(container) == "" {
        cn, err := kube.DetectSingleContainer(cfg.Namespace, pod)
        if err != nil { return err }
        container = cn
    }

    // 3) Compose object key
    ts := time.Now().Format("20060102150405")
    base := firstNonEmpty(cfg.BackupNamePref, cfg.WorkloadName, pod, "backup")
    key := fmt.Sprintf("%s-%s.tar.gz", sanitizeName(base), ts)
    s3URL := s3util.ObjectURL(cfg.AwsBucket, cfg.AwsPrefix, key)

    log.Printf("Start streaming archive at %s", time.Now().Format("02-01-2006 15:04:05"))

    // 4) Build pipeline: kubectl exec tar | aws s3 cp - s3://...
    tarCmd := exec.Command(
        "kubectl", "exec", "-n", cfg.Namespace, pod, "-c", container, "--",
        "tar", "czf", "-", "-C", cfg.SrcPath, ".",
    )
    args := []string{"s3", "cp", "-", s3URL, "--content-type", "application/gzip"}
    if strings.TrimSpace(cfg.AwsEndpoint) != "" {
        args = append([]string{"--endpoint-url", cfg.AwsEndpoint}, args...)
    }
    awsCmd := exec.Command("aws", args...)

    tarStdout, err := tarCmd.StdoutPipe()
    if err != nil { return fmt.Errorf("tar stdout: %w", err) }
    tarStderr := &bytes.Buffer{}
    tarCmd.Stderr = tarStderr

    awsCmd.Stdin = tarStdout
    awsStdout := &bytes.Buffer{}
    awsStderr := &bytes.Buffer{}
    awsCmd.Stdout = awsStdout
    awsCmd.Stderr = awsStderr

    if err := tarCmd.Start(); err != nil { return fmt.Errorf("start tar: %w", err) }
    if err := awsCmd.Start(); err != nil {
        _ = tarCmd.Process.Kill()
        return fmt.Errorf("start aws: %w", err)
    }
    tarErr := tarCmd.Wait()
    if c, ok := tarStdout.(io.Closer); ok { _ = c.Close() }
    awsErr := awsCmd.Wait()
    if tarErr != nil { return fmt.Errorf("tar: %v: %s", tarErr, tarStderr.String()) }
    if awsErr != nil { return fmt.Errorf("aws: %v: %s", awsErr, awsStderr.String()) }

    log.Printf("Upload completed at %s", time.Now().Format("02-01-2006 15:04:05"))

    // 5) TTL prune
    if strings.TrimSpace(cfg.BackupTTL) != "" {
        if err := s3util.PruneByTTL(cfg.AwsEndpoint, cfg.AwsBucket, cfg.AwsPrefix, cfg.BackupTTL); err != nil {
            log.Printf("warning: prune by TTL failed: %v", err)
        }
    }
    return nil
}
