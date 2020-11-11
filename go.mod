module github.com/TheCacophonyProject/thermal-recorder

go 1.12

require (
	github.com/TheCacophonyProject/event-reporter v1.3.2-0.20200210010421-ca3fcb76a231
	github.com/TheCacophonyProject/go-config v1.5.0
	github.com/TheCacophonyProject/go-cptv v0.0.0-20201110014247-4e8fc2c763ab
	github.com/TheCacophonyProject/lepton3 v0.0.0-20200909032119-e2b2b778a8ee
	github.com/TheCacophonyProject/window v0.0.0-20200312071457-7fc8799fdce7
	github.com/alexflint/go-arg v1.3.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.5.0 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b // indirect
	golang.org/x/sys v0.0.0-20201110211018-35f3e6cf4a65 // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	gopkg.in/yaml.v2 v2.2.8
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	periph.io/x/periph v3.6.4+incompatible
)

// We maintain a custom fork of periph.io at the moment.
replace periph.io/x/periph => github.com/TheCacophonyProject/periph v2.1.1-0.20200615222341-6834cd5be8c1+incompatible
