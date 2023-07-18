#!/bin/bash

systemctl daemon-reload

systemctl enable thermal-recorder.service
systemctl restart thermal-recorder.service
