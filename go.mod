module github.com/TheCacophonyProject/thermal-recorder

go 1.12

require (
	github.com/TheCacophonyProject/go-config v1.1.1
	github.com/TheCacophonyProject/go-cptv v0.0.0-20200121021233-067055d0edf0
	github.com/TheCacophonyProject/lepton3 v0.0.0-20200121020734-2ae28662e1bc
	github.com/TheCacophonyProject/window v0.0.0-20190821235241-ab92c2ee24b6
	github.com/alexflint/go-arg v1.1.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/nathan-osman/go-sunrise v0.0.0-20171121204956-7c449e7c690b // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.5.0 // indirect
	github.com/stretchr/testify v1.3.0
	golang.org/x/sys v0.0.0-20191117211529-81af7394a238 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	gopkg.in/yaml.v2 v2.2.5
	periph.io/x/periph v3.6.2+incompatible
)

// We maintain a custom fork of periph.io at the moment.
replace periph.io/x/periph => github.com/TheCacophonyProject/periph v2.0.1-0.20171123021141-d06ef89e37e8+incompatible
