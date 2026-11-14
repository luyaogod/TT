######
# Debug
######

# Prompt how to work
if [ "$1" = "" ] ; then
   echo "功能及用法說明:"
   echo "1.debug 程式              r.d program_name "
   echo "2.排程專用 debug 程式     r.d sch program_name "
   echo "Usage: "
   echo "1.debug program           r.d program_name "
   echo "2.schedule debug          r.d sch program_name "
   exit 1
fi

#版本查看
if [ "${1}" = '-v' ] ; then
   $FGLRUN $FBIN/42m/gendbg ${1}
   exit 0
fi

PW=`pwd`
mod=`pwd|awk -F/ '{print $NF}'`

# Setting source file path, and run the application in debug mode
FGLDEBPATH=$TOP/utl/3rd/fgldeb;export FGLDEBPATH
echo " "
echo " 即将执行 T100 专用版除错器, 如果发生任何问题无法执行,或希望使用原厂功能"
echo "        请改执行 'r.dg program_id'  "
echo " "
echo " 即將執行 T100 專用版除錯器, 如果發生任何問題無法執行,或希望使用原廠功能" 
echo "        請改執行 'r.dg program_id'  "
echo " "
FGLLDPATH=$FGLDEBPATH:$FGLDIR/lib:../42m:$FGLLDPATH;export FGLLDPATH
DBPATH=$DBPATH:$FGLDEBPATH;export DBPATH
FGLRESOURCEPATH=$FGLDEBPATH;export FGLRESOURCEPATH
FGLSOURCEPATH=../42m:$COM/lib/42m:$COM/sub/42m:$COM/qry/42m;export FGLSOURCEPATH
#add by hj
#$TEMPDIR/fgldeb
FGLRESOURCEPATH=$TEMPDIR/fgldeb
#是否為空
if [ ! -s $FGLRESOURCEPATH ] ; then
   mkdir $FGLRESOURCEPATH
   #cp $FGLDEBPATH/fgldeb.??? $FGLRESOURCEPATH      #171120-00018#1 mark
   #cp $FGLDEBPATH/*.42f $FGLRESOURCEPATH           #171120-00018#1 mark
fi
cp $FGLDEBPATH/fgldeb.??? $FGLRESOURCEPATH          #171120-00018#1 add   
cp $FGLDEBPATH/*.42f $FGLRESOURCEPATH               #171120-00018#1 add

cf1=`echo $1|cut -c1-3`

if [ ${cf1} = 'adz' ]
   then
    if [ -s ../42s/zh_TW/${1}.42s ] ; then
       cp -f ../42s/zh_TW/${1}.42s $FGLRESOURCEPATH/.
       chmod 777 $FGLRESOURCEPATH/${1}.42s
    fi
fi

FGLRUN=${FGLRUN:=fglrun};export FGLRUN
cd $FGLRESOURCEPATH

if [ ${cf1} = 'adz' ]
  then
    cmd="$FGLRUN $FGLDEBPATH/fgldeb $PW/../42r/${1}.42r BBDL512840855a "
  else
    #cmd="$FGLRUN $FBIN/42m/gendbg ${1} "     #171005-00011   mark
    #加入排程專用debug  {1}:sch {2}:作業編號
    cmd="$FGLRUN $FBIN/42m/gendbg ${1} ${2}"  #171005-00011   add
fi

i=2

while [ $i -le $# ]
do
 eval arg=\${$i}
echo ${arg}
 cmd="${cmd} '${arg}'"
 i=`expr $i + 1`
done
echo $cmd
eval $cmd

#$cmd "$@"

cd $PW

