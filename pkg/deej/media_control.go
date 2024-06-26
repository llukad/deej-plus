package deej

import (
    "fmt"
    "github.com/micmonay/keybd_event"
)

func PlayPause() error{
	// Create a new keyboard
    kb, err := keybd_event.NewKeyBonding()
    if err != nil {
        return fmt.Errorf("Error creating new keyboard for play/pause: %w", err)
    }

    // Play/Pause
    kb.SetKeys(keybd_event.VK_MEDIA_PLAY_PAUSE)
    err = kb.Launching()
    if err != nil {
        return fmt.Errorf("Error simulating Play/Pause key: %w", err)
    }

    return nil
}

func PrevTrack() error{
	// Create a new keyboard
    kb, err := keybd_event.NewKeyBonding()
    if err != nil {
        return fmt.Errorf("Error creating new keyboard for PrevTrack: %w", err)
    }

    // Play/Pause
    kb.SetKeys(keybd_event.VK_MEDIA_PREV_TRACK)
    err = kb.Launching()
    if err != nil {
        return fmt.Errorf("Error simulating PrevTrack key: %w", err)
    }

    return nil
}

func NextTrack() error{
	// Create a new keyboard
    kb, err := keybd_event.NewKeyBonding()
    if err != nil {
        return fmt.Errorf("Error creating new keyboard for NextTrack: %w", err)
    }

    // Play/Pause
    kb.SetKeys(keybd_event.VK_MEDIA_NEXT_TRACK)
    err = kb.Launching()
    if err != nil {
        return fmt.Errorf("Error simulating NextTrack key: %w", err)
    }

    return nil
}