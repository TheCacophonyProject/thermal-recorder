#!/bin/bash

systemctl daemon-reload

systemctl enable thermal-recorder.service
systemctl enable leptond.service
systemctl stop leptond.service
systemctl restart thermal-recorder.service
sleep 1
systemctl start leptond.service
