module github.com/TheCacophonyProject/thermal-recorder

go 1.12

require (
	github.com/TheCacophonyProject/go-cptv v0.0.0-20190522012101-cfce3277dd0d
	github.com/TheCacophonyProject/lepton3 v0.0.0-20190715012645-46e70a4e1b3f
	github.com/TheCacophonyProject/window v0.0.0-20180925042134-4e43785eca90
	github.com/alexflint/go-arg v0.0.0-20180516182405-f7c0423bd11e
	github.com/coreos/go-systemd v0.0.0-20180511133405-39ca1b05acc7
	github.com/felixge/pidctrl v0.0.0-20160307080219-7b13bcae7243
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/nathan-osman/go-sunrise v0.0.0-20171121204956-7c449e7c690b
	github.com/stretchr/testify v1.2.2
	gopkg.in/yaml.v2 v2.2.1
	periph.io/x/periph v0.0.0-00010101000000-000000000000
)

// We maintain a custom fork of periph.io at the moment.
replace periph.io/x/periph => github.com/TheCacophonyProject/periph v2.0.1-0.20171123021141-d06ef89e37e8+incompatible
