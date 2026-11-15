#!/bin/ksh
#231208-00026#1 2023/12/08 By 20992 因應flm升級到6.00.16 所以須對r.c r.cs r.f r.s 進行修正 才能使用對應的fgllicense

#授權:FLM6版本用FGLLICENSE,5.XX用FGLPROFILE
FGLPROFILE=$FGLDIR/etc/fglprofile.dev:$FGLPROFILE;export FGLPROFILE
#231208-00026#1 cancel mark
FGLLICENSE=$FGLDIR/etc/fgllicense.dev;export FGLLICENSE

#若為DM,僅此處放開下方TOPDB,其他地方維持ORA設定(所以DMG的profile也維持TOPDB=ORA)
#TOPDB=DMG; export TOPDB

if [ "${T100CTYPE}" != "en_US" ] ; then
   LC_CTYPE=zh_CN.utf8;export LC_CTYPE
   LC_COLLATE=zh_CN.utf8;export LC_COLLATE
fi

if [ "${TOPDB}" == "PGS" ] ; then
   echo INFO:以PostgreSQL版本編譯
   CDB=is_pgs
elif [ "${TOPDB}" == "DMG" ] ; then
   echo INFO:以 DM 版本編譯
   CDB=is_dmg
elif [ "${TOPDB}" == "MSV" ] ; then
   echo INFO:以MS-SQL Server版本編譯
   CDB=is_msv
else
   echo INFO:以ORACLE版本編譯
   CDB=is_ora
fi

if [ "${GENERODIR}" != "" ] ; then
   export LIBDIR=$GENERODIR/gre/lib:$GENERODIR/gst/bin:$GENERODIR/gst/lib
   PATH=$LIBDIR:$FGLDIR/bin:$PATH
else
   PATH=$FGLDIR/bin:$PATH 
fi
export PATH 

umask 000
killfulltype ${1}.4gl >/dev/null 2>&1
if [ "${1}" = "" ] ; then
   echo "用法: 一般小工具編譯 "
   echo "      r.cs source_code_filename"
   echo "      加入include檔案 ($TOP/com/inc) "
   echo "      r.cs -i source_code_filename "
   echo "      列出4GL內所有靜態的SQL "
   echo "      r.cs -S source_code_filename "
   exit
elif [ "${1}" = "-i" ] ; then
   fglcomp -D ${CDB} -I $TOP/com/inc --timestamp -W all ${2}

elif [ "${1}" = "-S" ] ; then
   fglcomp -D ${CDB} -I $TOP/com/inc --timestamp -W all -S ${2}
else
   GSLINTRULES=/u1/etc3/default.gslint;export GSLINTRULES
   GSLINTENABLED=1

   fglcomp -D ${CDB} --timestamp -W all ${1}
fi

