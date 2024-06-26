package deej

import (
    "fmt"
    // "golang.org/x/sys/windows"
    // "golang.org/x/sys/windows/svc"
    "golang.org/x/sys/windows/svc/mgr"
    // "os"
    "time"
    "os/exec"
    "strings"
)

func startService(serviceName string, configPath string) error {
	scm, err := mgr.Connect()
    if err != nil {
        return fmt.Errorf("failed to connect to SCM: %v", err)
    }
    defer scm.Disconnect()

	exists, err := serviceExists(scm, serviceName)
	if err != nil{
		return err
	}

	running := false

	if exists{
		service, err := scm.OpenService(serviceName)
	    if err != nil {
	        return err
	    }
    	defer service.Close()
		running, err = isSreviceRunning(service)
		if err != nil{
			return err
		}

		if !running{
			err = service.Start()
			if err != nil {
				return fmt.Errorf("Cant start service: %w", err)
			}
			err = waitUntilStarted(service)
			if err != nil {
				return err
			}else{
				return nil
			}
		}
	}

	if !exists {
		err = createTunnel(configPath)
		if err != nil{
			return fmt.Errorf("Error creating a tunnel: %w", err)
		}
		err = waitUntilCreated(scm, serviceName)
		if err != nil {
			return err
		}
		service, err := scm.OpenService(serviceName)
	    if err != nil {
	        return err
	    }
    	defer service.Close()

		err = waitUntilStarted(service)
		if err != nil {
			return err
		}else{
			return nil
		}
	}
	return nil
}

func stopService(serviceName string) error {
    scm, err := mgr.Connect()
    if err != nil {
        return fmt.Errorf("failed to connect to SCM: %v", err)
    }
    defer scm.Disconnect()

    exists, err := serviceExists(scm, serviceName)
	if err != nil{
		return err
	}

	if !exists{
		return nil
	}

	if exists{
		// service, err := scm.OpenService(serviceName)
	    // if err != nil {
	    //     return err
	    // }
    	// defer service.Close()
		// running, err := isSreviceRunning(service)
		// if err != nil{
		// 	return err
		// }

		// if !running{
		// 	return nil
		// }

		// if running{
		// 	_, err := service.Control(svc.Stop)
		//     if err != nil {
		//         return fmt.Errorf("failed to stop service %s: %v", serviceName, err)
		//     }
		//     err = waitUntilStopped(service)
		//     if err != nil {
		//     	return err
		//     }
		//     return nil
		// }

		err = deleteTunnel(strings.Split(serviceName, "$")[1])
		if err != nil {
			return err
		}

	}
    return nil
}


func serviceExists(scm *mgr.Mgr, serviceName string) (bool, error){
    service, err := scm.OpenService(serviceName)
    if err != nil {
        if err.Error() == "The specified service does not exist as an installed service."{
        	return false, nil
        }
    }
    defer service.Close()

    return true, nil
}

func isSreviceRunning(service *mgr.Service) (bool, error){
	status, err := service.Query()
    if err != nil {
        return false, fmt.Errorf("failed to get status service: %v", err)
    }

    if status.State == 4{
    	return true, nil
    }else if status.State == 1{
    	return false, nil
    }
    return false, nil
}


func createTunnel(configPath string) error{
	cmd := exec.Command("C:\\Program Files\\WireGuard\\wireguard.exe", "/installtunnelservice", configPath)
    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("Error while adding VPN tunnel %s: %w", configPath, err)
    }

    return nil
}

func deleteTunnel(serviceName string) error {
	cmd := exec.Command("C:\\Program Files\\WireGuard\\wireguard.exe", "/uninstalltunnelservice", serviceName)
    err := cmd.Start()
    if err != nil {
        return fmt.Errorf("Error while deleting VPN tunnel %s: %w", serviceName, err)
    }

    return nil
}

func waitUntilCreated(scm *mgr.Mgr, serviceName string) error {
	for{
		exists, err := serviceExists(scm, serviceName)
		if err != nil {
			return err
		}
		if exists{
			return nil
		}
		time.Sleep(100*time.Millisecond)
	}
}

func waitUntilStarted(service *mgr.Service) error{
	for{
		running, err := isSreviceRunning(service)
		if err != nil {
			return err
		}
		if running{
			return nil
		}
		time.Sleep(100*time.Millisecond)
	}
}

func waitUntilStopped(service *mgr.Service) error{
	for{
		running, err := isSreviceRunning(service)
		if err != nil {
			return err
		}
		if !running{
			return nil
		}
		time.Sleep(100*time.Millisecond)
	}
}

func StartVPN(serviceName string, configPath string) error{
	err := startService(serviceName, configPath)
    if err != nil {
    	return fmt.Errorf("Error starting VPN: %w", err)
    }
    return nil
}

func StopVPN(serviceName string) error{
	err := stopService(serviceName)
    if err != nil {
    	return fmt.Errorf("Error stopping VPN: %w", err)
    }
    return nil
}
 