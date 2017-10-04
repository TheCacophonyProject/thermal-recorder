# Setting up a Raspberry Pi 3 for thermal-recorder

* Start with Stretch Lite image
* Set a large SPI buffer
  - Add `spidev.bufsiz=65536` to /boot/cmdline.txt
* Setup up Raspberry Pi hardware
  - Enable SPI & I2C
  - Reduce GPU memory size to 16MB
  - Limit CPU and GPU core clocks for more stable SPI

Use this /boot/config.txt:

```
# For more options and information see
# http://rpf.io/configtxt
# Some settings may impact device functionality. See link above for details
# Additional overlays and parameters are documented /boot/overlays/README

gpu_mem=16
start_x=0
enable_uart=0

arm_freq=600
arm_freq_min=600

gpu_freq=200
gpu_freq_min=200

dtparam=i2c_arm=on
dtparam=spi=on

# Enable audio (loads snd_bcm2835)
dtparam=audio=on
```
