#!/bin/bash

systemctl daemon-reload

systemctl enable thermal-recorder.service
systemctl enable leptond.service
systemctl restart leptond.service
systemctl restart thermal-recorder.service
