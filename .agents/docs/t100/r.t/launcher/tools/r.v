#!/bin/ksh

if [ "$1" = "" ] ; then
   echo "Usage: r.v form_name language(0,1,2) customerlized(C/c/P/p) \n"
   exit 1
fi

#避免在topo底下執行 r.v 時,會抓取錯誤的42f,所以先改變執行目錄.
cd $AZZ/42m
FGLRESOURCEPATH=$TEMPDIR; export FGLRESOURCEPATH 
$FGLRUN azz_azzp189.42m $1 $2 $3 $4 $5
