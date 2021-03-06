// +build linux,!cgo

package serial

import (
	"os"
	"syscall"
	"time"
	"unsafe"
	"fmt"
)

const TCFLSH = 0x540B

func openPort(name string, baud int, readTimeout time.Duration) (p *Port, err error) {
	fmt.Println("linux openPort")

	var bauds = map[int]uint32{
		50:      syscall.B50,
		75:      syscall.B75,
		110:     syscall.B110,
		134:     syscall.B134,
		150:     syscall.B150,
		200:     syscall.B200,
		300:     syscall.B300,
		600:     syscall.B600,
		1200:    syscall.B1200,
		1800:    syscall.B1800,
		2400:    syscall.B2400,
		4800:    syscall.B4800,
		9600:    syscall.B9600,
		19200:   syscall.B19200,
		38400:   syscall.B38400,
		57600:   syscall.B57600,
		115200:  syscall.B115200,
		230400:  syscall.B230400,
		460800:  syscall.B460800,
		500000:  syscall.B500000,
		576000:  syscall.B576000,
		921600:  syscall.B921600,
		1000000: syscall.B1000000,
		1152000: syscall.B1152000,
		1500000: syscall.B1500000,
		2000000: syscall.B2000000,
		2500000: syscall.B2500000,
		3000000: syscall.B3000000,
		3500000: syscall.B3500000,
		4000000: syscall.B4000000,
	}

	rate := bauds[baud]

	if rate == 0 {
		return
	}

	f, err := os.OpenFile(name, syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0666)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil && f != nil {
			f.Close()
		}
	}()

	fd := f.Fd()
	//vmin, vtime := posixTimeoutValues(readTimeout)
	t := syscall.Termios{
		Iflag:  syscall.IGNPAR | syscall.ICRNL,
		Cflag:  syscall.CS8 | syscall.CREAD | syscall.CLOCAL | rate,
		Oflag: 0,
		Lflag: syscall.ICANON,
		Cc:     [32]uint8{
			syscall.VINTR: 0,
			syscall.VQUIT: 0,
			syscall.VERASE: 0,
			syscall.VKILL: 0,
			syscall.VEOF: 4,
			syscall.VTIME: 0,
			syscall.VMIN: 1,
			syscall.VSWTC: 0,
			syscall.VSTART: 0,
			syscall.VSTOP: 0,
			syscall.VSUSP: 0,
			syscall.VEOL: 0,
			syscall.VREPRINT: 0,
			syscall.VDISCARD: 0,
			syscall.VWERASE: 0,
			syscall.VLNEXT: 0,
			syscall.VEOL2: 0},
		Ispeed: rate,
		Ospeed: rate,
	}

	if _, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(TCFLSH),
		uintptr(syscall.TCIOFLUSH),
	); errno != 0 {
		return nil, errno
	}


	if _, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&t)),
		0,
		0,
		0,
	); errno != 0 {
		return nil, errno
	}

	if err = syscall.SetNonblock(int(fd), false); err != nil {
		return
	}

	return &Port{f: f}, nil
}

type Port struct {
	// We intentionly do not use an "embedded" struct so that we
	// don't export File
	f *os.File
}

func (p *Port) Read(b []byte) (n int, err error) {
	return p.f.Read(b)
}

func (p *Port) Write(b []byte) (n int, err error) {
	return p.f.Write(b)
}

// Discards data written to the port but not transmitted,
// or data received but not read
func (p *Port) Flush() error {
	_, _, err := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(p.f.Fd()),
		uintptr(TCFLSH),
		uintptr(syscall.TCIOFLUSH),
	)
	return err
}

func (p *Port) Close() (err error) {
	return p.f.Close()
}
