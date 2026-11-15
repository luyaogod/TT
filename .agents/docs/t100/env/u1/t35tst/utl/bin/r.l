######
#Link execute binary
######

umask 000
export TOPENT=99
LC_CTYPE=zh_TW.utf8;export LC_CTYPE
LC_COLLATE=zh_TW.utf8;export LC_COLLATE

# Prompt how to work
if [ "$1" = "" ] ; then
   echo "功能及用法說明:"
   echo "  1.產出模組42x  r.l 模組編號           例如 r.l axm 或 r.l sub "
   echo "  2.產出程式42r  r.l 主程式編號         例如 r.l axmt500 "
   echo "  3.重新產出有連結到指定子程式的所有42r 例如 r.l azzi800_01 ALL "
   echo "  4.指定模組下的所有主程式重新連結      例如 r.l axm ALL "
   echo "  5.協助檢查目錄的環境設定與42x是否正常 例如 r.l -check azzi900 "
   echo " "
   echo "Usage:"
   echo "  1.Generate composed 42x file                     r.l module-id"
   echo "  2.Generate 42r file for execute                  r.l program-id"
   echo "  3.Re-generate all 42r including assigned sub-4gl r.l azzi800_01 ALL"
   echo "  4.Re-generate all 42r in assigned module         r.l module-id ALL"
   echo "  5.Check environment and 42x for assigned 42r     r.l -check program-id"
   exit 1
fi        
         
# 連結作業
$FGLRUN $FBIN/42m/gen42r.42r $1 $2

