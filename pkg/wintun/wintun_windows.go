package wintun

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	
	"golang.zx2c4.com/wireguard/tun"
)

const (
	ExitSetupSuccess = 0
	ExitSetupFailed  = 1
)

func InstallWinTUN(interfaceName string) (newinterfaceName string, result bool) {
	logger := device.NewLogger(
		device.LogLevelVerbose,
		fmt.Sprint("(%s) ", interfaceName),
	)
	tun, err := tun.CreateTUN(interfaceName, 0)
	if err != nil {
		logger.Errorf("Failed to create TUN device:%v", err)
		return interfaceName, false
	}
	realInterfaceName, err2 := tun.Name()
	if err2 == nil {
		newinterfaceName = realInterfaceName
	}
	device := device.NewDevice(tun, conn.NewDefaultBind(), logger)
	err = device.Up()
	if err != nil {
		logger.Errorf("failed to bring up device:%v", err)
		return newinterfaceName, false
	}
	uapi, err := ipc.UAPIListen(newinterfaceName)
	if err != nil {
		logger.Errorf("failed to listen on uapi socket:%v", err)
		return newinterfaceName, false
	}
	errs := make(chan error)
	term := make(chan os.Signal, 1)

	go func() {
		for {
			conn, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go device.IpcHandle(conn)
		}
	}()
	logger.Verbosef("UAPI listener started")
	signal.Notify(term, os.Interrupt)
	signal.Notify(term, os.Kill)
	signal.Notify(term, syscall.SIGTERM)
	select {
	case <-term:
	case <-errs:
	case <-device.Wait():
	}
	return newinterfaceName, true
}
