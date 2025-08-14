package config

import (
    "fmt"
    "os"
    "strings"
)

type Config struct {
    Namespace       string
    WorkloadKind    string
    WorkloadName    string
    LabelSelector   string
    Container       string
    SrcPath         string
    BackupNamePref  string

    AwsEndpoint     string
    AwsBucket       string
    AwsPrefix       string
    AwsRegion       string
    AwsAccessKeyID  string
    AwsSecret       string

    BackupTTL       string
}

func getenv(key, def string, env map[string]string) string {
    if v, ok := env[key]; ok && v != "" {
        return v
    }
    return def
}

func FromEnv(envList []string) (Config, error) {
    env := map[string]string{}
    for _, kv := range envList {
        parts := strings.SplitN(kv, "=", 2)
        if len(parts) == 2 {
            env[parts[0]] = parts[1]
        }
    }
    // Namespace: prefer KUBE_NAMESPACE, then NAMESPACE, default to "default"
    ns := getenv("KUBE_NAMESPACE", getenv("NAMESPACE", "default", env), env)
    cfg := Config{
        Namespace:      ns,
        WorkloadKind:   strings.ToLower(getenv("WORKLOAD_KIND", "", env)),
        WorkloadName:   getenv("WORKLOAD_NAME", "", env),
        LabelSelector:  getenv("POD_LABEL_SELECTOR", "", env),
        Container:      getenv("CONTAINER", "", env),
        SrcPath:        getenv("SRC_PATH", "", env),
        BackupNamePref: getenv("BACKUP_NAME_PREFIX", "", env),

        AwsEndpoint:    getenv("AWS_S3_ENDPOINT_URL", "", env),
        AwsBucket:      getenv("AWS_BUCKET_NAME", "", env),
        AwsPrefix:      getenv("AWS_BUCKET_BACKUP_PATH", "", env),
        AwsRegion:      getenv("AWS_DEFAULT_REGION", "", env),
        AwsAccessKeyID: getenv("AWS_ACCESS_KEY_ID", os.Getenv("AWS_ACCESS_KEY_ID"), env),
        AwsSecret:      getenv("AWS_SECRET_ACCESS_KEY", os.Getenv("AWS_SECRET_ACCESS_KEY"), env),

        BackupTTL:      getenv("BACKUP_TTL", "", env),
    }
    // Validate required
    missing := []string{}
    if strings.TrimSpace(cfg.AwsBucket) == "" { missing = append(missing, "AWS_BUCKET_NAME") }
    if strings.TrimSpace(cfg.AwsRegion) == "" { missing = append(missing, "AWS_DEFAULT_REGION") }
    if strings.TrimSpace(cfg.AwsAccessKeyID) == "" { missing = append(missing, "AWS_ACCESS_KEY_ID") }
    if strings.TrimSpace(cfg.AwsSecret) == "" { missing = append(missing, "AWS_SECRET_ACCESS_KEY") }
    if strings.TrimSpace(cfg.SrcPath) == "" { missing = append(missing, "SRC_PATH") }
    if len(missing) > 0 {
        return cfg, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
    }
    return cfg, nil
}
