#!/bin/sh
systemctl daemon-reload
systemctl enable goconnect-daemon
echo "GoConnect Daemon installed. Start with: sudo systemctl start goconnect-daemon"
