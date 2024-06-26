package deej

import (
    "fmt"
    "github.com/gek64/displayController"
    "time"
)

func ChangeBrightness(monitorID int, brightness int) error{
    // Get the system display devices
    compositeMonitors, err := displayController.GetCompositeMonitors()
    if err != nil {
        return fmt.Errorf("Error getting system monitors: %w", err)
    }

    err = displayController.SetVCPFeature(compositeMonitors[monitorID].PhysicalInfo.Handle, 0x10, brightness)
    if err != nil {
        return fmt.Errorf("Error setting monitor brightness: %w", err)
    }

    return nil
}

func ChangeEyeProtection(monitorID int, level int, brightness int) error{
    // Get the system display devices
    compositeMonitors, err := displayController.GetCompositeMonitors()
    if err != nil {
        return fmt.Errorf("Error getting system monitors: %w", err)
    }

    err = displayController.SetVCPFeature(compositeMonitors[monitorID].PhysicalInfo.Handle, 0xE6, level)
    if err != nil {
        return fmt.Errorf("Error setting monitor brightness: %w", err)
    }

    time.Sleep(500*time.Millisecond)

    err = displayController.SetVCPFeature(compositeMonitors[monitorID].PhysicalInfo.Handle, 0x10, brightness)
    if err != nil {
        return fmt.Errorf("Error setting monitor brightness: %w", err)
    }

    return nil
}

func ChangeColorMode(monitorID int, mode int, brightness int) error{
    // MODES for asus PROART PA279CV:
    // 00 / 0 - Standard
    // 0B / 11 - Scenery
    // 0D / 13 - sRGB
    // 21 / 33 - Rec. 709
    // 23 / 35 - DCI-P3
    compositeMonitors, err := displayController.GetCompositeMonitors()
    if err != nil {
        return fmt.Errorf("Error getting system monitors: %w", err)
    }

    err = displayController.SetVCPFeature(compositeMonitors[monitorID].PhysicalInfo.Handle, 0xDC, mode)
    if err != nil {
        return fmt.Errorf("Error setting monitor brightness: %w", err)
    }

    time.Sleep(500*time.Millisecond)

    err = displayController.SetVCPFeature(compositeMonitors[monitorID].PhysicalInfo.Handle, 0x10, brightness)
    if err != nil {
        return fmt.Errorf("Error setting monitor brightness: %w", err)
    }

    return nil
}

func SColorMode(monitorID int) error{
    // Get the system display devices
    compositeMonitors, err := displayController.GetCompositeMonitors()
    if err != nil {
        return fmt.Errorf("Error getting system monitors: %w", err)
    }

    currentValue, _, err := displayController.GetVCPFeatureAndVCPFeatureReply(compositeMonitors[monitorID].PhysicalInfo.Handle, 0x14)
    if err != nil {
        return fmt.Errorf("Error getting color mode: %w", err)
    }else{
        fmt.Print(currentValue)
    }

    return nil
}