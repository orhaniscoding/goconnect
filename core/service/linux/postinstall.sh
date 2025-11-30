#!/bin/sh
systemctl daemon-reload
systemctl enable goconnect-server
echo "GoConnect Server installed. Start with: sudo systemctl start goconnect-server"
