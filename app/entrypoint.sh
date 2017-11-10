#!/bin/sh
cp /usr/share/zoneinfo/$TZ /etc/localtime
echo $TZ > /etc/timezone

crond
/app/goapp