######
#產生程式藍圖與程式組合
######

umask 000
export DBDATE=Y4MD/

#若TOPENT為空則指定為99,否則使用原有設定
if [ "$TOPENT" = "" ] ; then
   export TOPENT=99
fi

#程式產生路徑判斷工具
$FGLRUN $FBIN/42m/compose_path $1 a:$2 b:$3 c:$4 d:$5

