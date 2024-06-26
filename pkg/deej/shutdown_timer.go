package deej

import (
    "fmt"
    "os/exec"
    "strconv"
)

func cancelShutdown() error {
    // Construct the shutdown command with the /a option
    cmd := exec.Command("shutdown", "/a")
    
    // Run the command
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to execute cancel shutdown command: %v", err)
    }
    
    return nil
}

func scheduleShutdown(seconds int) error {
    // Convert the seconds to a string
    secondsStr := strconv.Itoa(seconds)
    
    // Construct the shutdown command with the /s /t options
    cmd := exec.Command("shutdown", "/s", "/t", secondsStr)
    
    // Run the command
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to execute shutdown command: %v", err)
    }
    
    return nil
}

func ShutdownIn(seconds int) error {
    // Cancel any existing shutdown
    err := cancelShutdown()
    if err != nil {
        // fmt.Println("No existing shutdown to cancel or failed to cancel existing shutdown")
        // Schedule a new shutdown
        err2 := scheduleShutdown(seconds)
        if err2 != nil {
            return fmt.Errorf("failed to schedule shutdown: %v", err2)
        }
    } else {
        // fmt.Println("Canceled any existing shutdown")
    }


    return nil
}
