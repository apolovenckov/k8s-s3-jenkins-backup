package execx

import (
    "bytes"
    "fmt"
    "log"
    "os/exec"
    "strings"
)

func Run(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr
    log.Printf("exec: %s %s", name, strings.Join(args, " "))
    if err := cmd.Run(); err != nil {
        if stderr.Len() > 0 {
            return out.String(), fmt.Errorf("%w: %s", err, stderr.String())
        }
        return out.String(), err
    }
    return out.String(), nil
}

