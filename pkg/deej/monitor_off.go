package deej

import (
    "context"
    "fmt"
    "os/exec"
    "time"
)

func MonitorOff() error {
    // Create a new context with a timeout
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel() // Always call cancel to release resources associated with the context

    // Create the command
    cmd := exec.CommandContext(ctx, "powershell", "-Command", "(Add-Type '[DllImport(\"user32.dll\")]public static extern int SendMessage(int hWnd, int hMsg, int wParam, int lParam);' -Name a -Pas)::SendMessage(-1,0x0112,0xF170,2)")

    // Execute the command
    _, err := cmd.CombinedOutput()

    // Check if there was an error
    if ctx.Err() == context.DeadlineExceeded {
        //fmt.Println("Command timed out")
        return nil
    }

    // Handle other errors
    if err != nil {
        
        return fmt.Errorf("Command execution failed: %v", err)
    }
    return nil
}
