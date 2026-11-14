######
# Compiler source
######
#231208-00026#1 2023/12/08 By 20992 因應flm升級到6.00.16 所以須對r.c r.cs r.f
#r.s 進行修正 才能使用對應的fgllicense
umask 000
#unset GSTWCDIR

FGLPROFILE=$FGLDIR/etc/fglprofile.dev:$FGLPROFILE;export FGLPROFILE
#FGLDBPATH=$FGLDBPATH/ext; export FGLDBPATH
#231208-00026#1 add
FGLLICENSE=$FGLDIR/etc/fgllicense.dev;export FGLLICENSE 

#若有需要看中文報錯訊息,請將下列LC_CTYPE/LC_COLLATE指令放開mark
#LC_CTYPE=zh_TW.utf8;export LC_CTYPE
#LC_COLLATE=zh_TW.utf8;export LC_COLLATE
#繁體中文請輸入zh_TW / 簡體中文請輸入 zh_CN

# Prompt how to work
if [ "$1" = "" ] ; then
   echo "功能及用法說明:"
   echo "  1.編譯並會執行相關作業的42s檔重產 r.f screen_form_filename"
   echo "  2.編譯(不帶其他功能)              r.f screen_form_filename tiptop"
   echo " "
   echo "Usage:"
   echo "  1.Compile and generate releational 42s r.f screen_form_filename"
   echo "  2.Compile(without others)              r.f screen_form_filename tiptop"
   exit 1
fi

# 編譯作業
GP=$DEV_VER
$FGLRUN $FBIN/42m${GP}/gen42f "$@"

