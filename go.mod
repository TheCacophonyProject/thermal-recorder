module github.com/TheCacophonyProject/thermal-recorder

go 1.12

require (
	github.com/TheCacophonyProject/go-config v0.0.0-20190918041555-29df9aabd14b
	github.com/TheCacophonyProject/go-cptv v0.0.0-20190522012101-cfce3277dd0d
	github.com/TheCacophonyProject/lepton3 v0.0.0-20190715012645-46e70a4e1b3f
	github.com/TheCacophonyProject/window v0.0.0-20190821235241-ab92c2ee24b6
	github.com/alexflint/go-arg v1.1.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/felixge/pidctrl v0.0.0-20160307080219-7b13bcae7243
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.2.2
	periph.io/x/periph v0.0.0-00010101000000-000000000000
)

// We maintain a custom fork of periph.io at the moment.
replace periph.io/x/periph => github.com/TheCacophonyProject/periph v2.0.1-0.20171123021141-d06ef89e37e8+incompatible
