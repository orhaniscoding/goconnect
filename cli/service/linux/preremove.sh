#!/bin/sh
systemctl stop goconnect-daemon || true
systemctl disable goconnect-daemon || true
