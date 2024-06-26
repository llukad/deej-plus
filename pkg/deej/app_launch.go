package deej

import (
    "fmt"
    "os/exec"
)


func LaunchApp(appPath string) error {
    // Step 1: Launch the application
    cmd := exec.Command(appPath)
    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("Error while launching app %s: %w", appPath, err)
    }

    return nil
}

func OpenWebsite(url string) error {
    cmd := exec.Command("cmd", "/c", "start", url)
    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("Error while opening website %s: %w", url, err)
    }

    return nil
}

func ExecuteInTerminal(command string) error {
    cmd := exec.Command("wt", "cmd", "/c", command)
    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("Error while launching command in terminal %s: %w", command, err)
    }

    return nil
}
