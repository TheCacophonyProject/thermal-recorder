module github.com/TheCacophonyProject/thermal-recorder

go 1.15

require (
	github.com/TheCacophonyProject/event-reporter v1.3.2-0.20200210010421-ca3fcb76a231
	github.com/TheCacophonyProject/go-config v1.7.1-0.20220118192751-a7abeb2ad25e
	github.com/TheCacophonyProject/go-cptv v0.0.0-20211109233846-8c32a5d161f7
	github.com/TheCacophonyProject/lepton3 v0.0.0-20210324024142-003e5546e30f
	github.com/TheCacophonyProject/window v0.0.0-20200312071457-7fc8799fdce7
	github.com/alexflint/go-arg v1.4.2
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/godbus/dbus v4.1.0+incompatible
	github.com/juju/ratelimit v1.0.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	gopkg.in/yaml.v2 v2.4.0
	periph.io/x/periph v3.6.8+incompatible
)

// We maintain a custom fork of periph.io at the moment.
replace periph.io/x/periph => github.com/TheCacophonyProject/periph v2.1.1-0.20200615222341-6834cd5be8c1+incompatible

//replace github.com/TheCacophonyProject/go-config => ../go-config
