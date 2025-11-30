#!/bin/sh
systemctl stop goconnect-server || true
systemctl disable goconnect-server || true
