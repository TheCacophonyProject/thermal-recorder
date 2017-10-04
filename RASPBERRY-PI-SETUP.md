* Start with Stretch Lite image
* Using raspi-config
  - Enable SPI
  - Enable I2C
  - Reduce GPU memory size to 16MB
* Set a large SPI buffer
  - Add `spidev.bufsiz=65536` to /boot/cmdline.txt
* Limit the CPU and core clocks to for stable SPI in /boot/config.txt
```
    arm_freq=600
    arm_freq_min=600

    core_freq=200
    core_freq_min=200
```
* Use the `powersave` CPU governor
  - NOTE: suspect this is irrelevant with the CPU clock limited
  - apt install cpufrequtils
  - GOVERNOR="powersave" to /etc/default/cpufrequtils
