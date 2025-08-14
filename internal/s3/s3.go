package s3

import (
    "bufio"
    "fmt"
    "log"
    "strings"
    "time"

    "k8s-s3-backup/internal/execx"
)

func AWS(args ...string) (string, error) { return execx.Run("aws", args...) }

func ObjectURL(bucket, prefix, key string) string {
    b := strings.TrimSpace(bucket)
    p := strings.Trim(prefix, "/")
    if p == "" { return fmt.Sprintf("s3://%s/%s", b, key) }
    return fmt.Sprintf("s3://%s/%s/%s", b, p, key)
}

func ListURL(bucket, prefix string) string {
    b := strings.TrimSpace(bucket)
    p := strings.Trim(prefix, "/")
    if p == "" { return fmt.Sprintf("s3://%s/", b) }
    return fmt.Sprintf("s3://%s/%s/", b, p)
}

type Object struct {
    Key  string
    Time time.Time
}

// Parse `aws s3 ls` output
func ParseLsObjects(out string) []Object {
    res := []Object{}
    sc := bufio.NewScanner(strings.NewReader(out))
    for sc.Scan() {
        line := strings.TrimSpace(sc.Text())
        if line == "" { continue }
        f := strings.Fields(line)
        if len(f) < 4 { continue }
        dt := strings.Join(f[0:2], " ")
        t, err := time.ParseInLocation("2006-01-02 15:04:05", dt, time.Local)
        if err != nil { continue }
        key := f[len(f)-1]
        if key == "" || strings.HasSuffix(key, "/") { continue }
        res = append(res, Object{Key: key, Time: t})
    }
    return res
}

func ParseTTL(s string) (time.Duration, error) {
    s = strings.TrimSpace(s)
    if s == "" { return 0, fmt.Errorf("empty TTL") }
    if strings.HasSuffix(s, "d") {
        nStr := strings.TrimSuffix(s, "d")
        var n int
        if _, err := fmt.Sscanf(nStr, "%d", &n); err != nil { return 0, err }
        return time.Duration(n) * 24 * time.Hour, nil
    }
    return time.ParseDuration(s)
}

func PruneByTTL(endpoint, bucket, prefix, ttlStr string) error {
    d, err := ParseTTL(ttlStr)
    if err != nil { return fmt.Errorf("invalid BACKUP_TTL: %w", err) }
    cutoff := time.Now().Add(-d)
    listURL := ListURL(bucket, prefix)
    args := []string{"s3", "ls", listURL}
    if strings.TrimSpace(endpoint) != "" {
        args = append([]string{"--endpoint-url", endpoint}, args...)
    }
    out, err := AWS(args...)
    if err != nil { return err }
    objs := ParseLsObjects(out)
    for _, o := range objs {
        if o.Time.Before(cutoff) {
            del := []string{"s3", "rm", ObjectURL(bucket, prefix, o.Key)}
            if strings.TrimSpace(endpoint) != "" {
                del = append([]string{"--endpoint-url", endpoint}, del...)
            }
            if _, err := AWS(del...); err != nil {
                log.Printf("warning: failed to delete %s: %v", o.Key, err)
            }
        }
    }
    return nil
}
