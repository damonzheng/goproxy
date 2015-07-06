#!/bin/sh
#
# goagent init script
#

### BEGIN INIT INFO
# Provides:          goagent
# Required-Start:    $syslog
# Required-Stop:     $syslog
# Should-Start:      $local_fs
# Should-Stop:       $local_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: Monitor for goagent activity
# Description:       goagent is a gae proxy forked from gappproxy/wallproxy.
### END INIT INFO

# **NOTE** bash will exit immediately if any command exits with non-zero.
set -e

PACKAGE_NAME=goagent
PACKAGE_DESC="goagent proxy server"
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:${PATH}

start() {
    echo -n "Starting ${PACKAGE_DESC}: "
    nohup ./goproxy -v=2 /opt/goproxy/goproxy -addr=0.0.0.0:8000 -pidfile /var/run/goproxy.pid -v=1 -logtostderr=0 -log_dir=/var/log/ &
    echo "${PACKAGE_NAME}."
}

stop() {
    echo -n "Stopping ${PACKAGE_DESC}: "
    kill `cat /var/run/goproxy.pid` >/dev/null 2>&1 || true
    echo "${PACKAGE_NAME}."
}

reload() {
    echo -n "Reloading ${PACKAGE_DESC}: "
    kill -HUP `cat /var/run/goproxy.pid` >/dev/null 2>&1 || true
    echo "${PACKAGE_NAME}."
}

restart() {
    stop || true
    sleep 1
    start
}

usage() {
    N=$(basename "$0")
    echo "Usage: [sudo] $N {start|stop|reload|restart}" >&2
    exit 1
}

if [ "$(id -u)" != "0" ]; then
    echo "please use sudo to run ${PACKAGE_NAME}"
    exit 0
fi

# `readlink -f` won't work on Mac, this hack should work on all systems.
cd $(python -c "import os; print(os.path.dirname(os.path.realpath('$0')))")

case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    reload)
        reload
        ;;
    restart)
        restart
        ;;
    *)
        usage
        ;;
esac

exit 0
