module github.com/lBeJIuk/streamdeckd

go 1.18

require (
	github.com/gobwas/ws v1.1.0
	github.com/godbus/dbus/v5 v5.0.6
	github.com/shirou/gopsutil/v3 v3.21.9
	github.com/unix-streamdeck/api v1.0.1
	github.com/unix-streamdeck/driver v0.0.0-20211119182210-fc6b90443bcd
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
)

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/fogleman/gg v1.3.0 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/karalabe/hid v1.0.1-0.20190806082151-9c14560f9ee8 // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	golang.org/x/image v0.0.0-20211028202545-6944b10bf410 // indirect
	golang.org/x/sys v0.0.0-20210816074244-15123e1e1f71 // indirect
)

replace github.com/unix-streamdeck/api v1.0.1 => ../api/
