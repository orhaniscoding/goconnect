#!/bin/sh
systemctl daemon-reload
systemctl enable goconnect
echo "GoConnect Daemon installed. Start with: sudo systemctl start goconnect"
