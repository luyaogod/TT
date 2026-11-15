#!/bin/sh
if [ $# -lt 1 ] ; then
   echo "Usage: $0 username 姓名"
   exit
fi

today=`date +"%m%d"`
echo $1
echo $2
useradd -g topdev -G topprd -s /bin/ksh -d /u3/usr/"$1" -m -c "$2" $1

echo $1 | passwd --stdin $1
echo ". /u3/pub/bin/chenv" > /u3/usr/"$1"/.profile

passwd -e $1

echo "第一次登入請更改密碼!"
