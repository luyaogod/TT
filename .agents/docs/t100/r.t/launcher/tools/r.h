#!/bin/ksh
######
# Run application in SQL debug mode plus
# 20210113 BY 12731 : r.h PLUS version ( Genero SQL debug )
######
## 1. Create SQL Debug log
t100version=$ERPVER  #20210113 add
if [ -z $1 ] ; then
  echo "Usage                    : r.h prog_name"
  #20210113 add (s)
  if [ "$t100version" == "3.0" ] ; then
      echo "==================================  "
      echo "How to open plus version : step 1 : export SQLDEBUGPLUS=on"
      echo "                           use    : r.h prog_name"
  fi
  #20210113 add (e)
  exit
fi

export T100RUNWAIT=1
FGLSQLDEBUG=9; export FGLSQLDEBUG
cd $TEMPDIR
r.r $@ 2> $TEMPDIR/sqllog.$1.$$.log

cf1=`echo $1|cut -c1-3`

## 2. fgldebug
debug=$SQLDEBUGPLUS  #20210113 add
export T100RUNWAIT=0
FGLSQLDEBUG=0; export FGLSQLDEBUG

#$FGLRUN $UTL/fbin/42m/T100SQLDebug.42r sqllog.$1.$$.log &  #20210113 mark

# new_version

#20210113 add (s)
if [ "$t100version" == "3.0" ] ; then 
   if [ "$debug" == "on" ] ; then
      $FGLRUN $UTL/bin/sqldebug/fglsqldebug -f sqllog.$1.$$.log &
   fi
fi

$FGLRUN $UTL/fbin/42m/T100SQLDebug.42r sqllog.$1.$$.log &

#20210113 add (e)


# EOF
