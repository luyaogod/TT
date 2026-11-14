######
# Compiler source
######
#231208-00026#1 2023/12/08 By 20992 因應flm升級到6.00.16 所以須對r.c r.cs r.f r.s 進行修正 才能使用對應的fgllicense
umask 000

FGLPROFILE=$FGLDIR/etc/fglprofile.dev:$FGLPROFILE;export FGLPROFIE
#231208-00026#1 add
FGLLICENSE=$FGLDIR/etc/fgllicense.dev;export FGLLICENSE

#若有需要看中文報錯訊息,請將下列LC_CTYPE/LC_COLLATE指令放開mark
#LC_CTYPE=zh_TW.utf8;export LC_CTYPE
#LC_COLLATE=zh_TW.utf8;export LC_COLLATE
#繁體中文請輸入zh_TW / 簡體中文請輸入 zh_CN

# Prompt how to work
if [ "$1" = "" ] ; then
   echo "功能及用法說明:"
   echo "  1.編譯                         r.c source_code_filename "
   echo "  2.編譯(不報錯)                 r.c source_code_filename 'tiptop' "
   echo "  3.編譯(TEMPDIR中進行臨時驗證)  r.c source_code_filename TEMP     "
   echo "  4.編譯用到指定表的所有4gl      r.c single_table_name    'ALL' "
   echo "  5.編譯附帶產出rdd檔案          r.c source_code_filename 'rdd' "
   echo "  "
   echo "Usage:"
   echo "  1.Compile source code             r.c source_code_filename "
   echo "  2.Compile without log             r.c source_code_filename 'tiptop' "
   echo "  3.Compile in TEMPDIR for test     r.c source_code_filename TEMP     "
   echo "  4.Compile select target by table  r.c single_table_name    'ALL' "
   echo "  5.Generate rdd file               r.c source_code_filename 'rdd' "
   exit 1
fi

#全形空白的處理, 工具為Genero 3.00以上版本,請放開下面 killfulltype
#killfulltype ${1}.4gl >/dev/null 2>&1
#編譯程式
GP=$GENPATH
if [ ! "${GP}" == "" ]; then
   echo "INFO:變數GENPATH存在,改變gen42m讀取路徑至 ${FBIN}/42m${GP}"
fi

$FGLRUN $FBIN/42m${GP}/gen42m ${1} ${2}

