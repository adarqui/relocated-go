#!/bin/bash -x

export PATH=/bin:/usr/bin/:/sbin:/usr/sbin

echo "`date`: $@" >> /tmp/all.log

#if [ $# -lt 4 ] ; then
#	exit -1
#fi

XNAME="$1"
XNAMESPACE="$2"
XCLASS="$3"
XPATH="$4"
XDESTINATION="$5"

echo "copying: [${XPATH}] to [${XDESTINATION}]" >> /tmp/all.log

mv "${XPATH}" "${XDESTINATION}"

exit 0
