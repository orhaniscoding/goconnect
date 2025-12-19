#!/bin/sh
systemctl stop goconnect || true
systemctl disable goconnect || true
