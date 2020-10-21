module github.com/TheCacophonyProject/thermal-recorder

go 1.12

require (
	github.com/TheCacophonyProject/event-reporter v1.3.2-0.20200210010421-ca3fcb76a231
	github.com/TheCacophonyProject/go-config v1.5.0
	github.com/TheCacophonyProject/go-cptv v0.0.0-20200928004402-97185afb2458
	github.com/TheCacophonyProject/lepton3 v0.0.0-20200909032119-e2b2b778a8ee
	github.com/TheCacophonyProject/window v0.0.0-20200312071457-7fc8799fdce7
	github.com/alexflint/go-arg v1.1.0
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.5.0 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/net v0.0.0-20200927032502-5d4f70055728 // indirect
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	gopkg.in/yaml.v2 v2.2.8
	periph.io/x/periph v3.6.4+incompatible
)

replace github.com/TheCacophonyProject/go-cptv => ../go-cptv

// We maintain a custom fork of periph.io at the moment.
replace periph.io/x/periph => github.com/TheCacophonyProject/periph v2.1.1-0.20200615222341-6834cd5be8c1+incompatible
