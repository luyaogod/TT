######
# Generating database schema for Database (ORA/PGS-MOG-/DMG)
######

umask 000
#ver="1.01.02 build-20160913" #原始版本
#ver="1.1.03 build-20170412"  #調整多語言說明及報錯
#ver="1.1.04 build-20171221"  #調整cat 產生到 $TOP/sch/ext/ds.sch 段落 #171221-00029#1
#ver="1.1.05 build-20180206"  #調整cat 增加刪除rm -rf $TOP/sch/ext/ds.sch 段落 #180206-00011#1
#ver="1.1.06 build-20181225"  #有調整顯示背景訊息順序,產生 sch/ext/ds.sch 後顯示訊息 #181225-00003#1
#ver="1.1.07 build-20190919"  #因應Datetime控件, 還原原本特別加工的段落 #190924-00043#1
#ver="1.1.08 build-20220712"  #sch/ext目錄下的ds.42d可能會有壓不掉的狀況,更新過程中強制刪除
#ver="1.1.09 build-20221018"  #表格名稱出現 $號刪除
#ver="1.1.10 build-20221215"  #匯出資料濾除test/temp開頭的表格名稱
#ver="1.1.11 build-20230510"  #reload前方的FGLPROFILE重設指令錯置
#ver="1.1.12 build-20230829"  #增加reload_gztz/reload_partial版本識別
#ver="1.1.13 build-20231211"  #231208-00026#1 2023/12/08 By 20992 因應flm升級到6.00.16 所以須對r.c r.cs r.f r.s 進行修正 才能使用對應的fgllicense
ver="1.1.14 build-20240919"   #240919-00048#1 2024/09/19 By 00824 3.0版DateTimeEdit元件開日期窗點選日期後，出現格式錯誤


# Prompt how to work
if [ "$1" = "" ] ; then
   echo "功能及用法說明:"
   echo "  1.重新產生指定db的schema檔案   r.s db_name"
   echo "  2.在更新指定表格的schema部分   r.s db_name  table_name"
   echo "  "
   echo "Usage:"
   echo "  1.Re-generate a schema file for assigned db. r.s db_name"
   echo "  2.Update assigned table in schema file.      r.s db_name table_name"
   exit 1
fi

#版次
if [ "$1" = "-v" ] ; then
   echo "r.s    版本為:"$ver
   $FGLRUN $FBIN/42m/gensch.42r -v
   $FGLRUN $FBIN/42m/reload_gztz.42r -v
   $FGLRUN $FBIN/42m/reload_partial.42m -v
   exit 1
fi
 
# Change to schema file directory, and delete the previous one
PW=`pwd`
FGLRUN=${FGLRUN:=fglrun};export FGLRUN 
FGLPROFILE=$FGLDIR/etc/fglprofile.dev:$FGLPROFILE;export FGLPROFILE
#231208-00026#1 cancel mark
FGLLICENSE=$FGLDIR/etc/fgllicense.dev;export FGLLICENS

DBDELIMITER=^;export DBDLIMITER

$FGLRUN $FBIN/42m/gensch.42r ${1} ${2}
if [ $? -ne 0 ] ; then
   err="err"
else
   if [ "$1" = "ds" ] ; then
      #補上alm的表格資料,便於編譯
      $TOP/utl/bin/reload_alm.sh ${1}
   
      #去除tt開頭的表資料
      sed -e '/^tt/d'         $TOP/sch/ds.sch    > $TOP/sch/ds.t.sch1
      #去除_t_的表資料
      sed -e '/_t_/d'         $TOP/sch/ds.t.sch1 > $TOP/sch/ds.t.sch2
      #去除 erp_ 開頭的 view 資料   #2017/04/11
      sed -e '/^erp_/d'       $TOP/sch/ds.t.sch2 > $TOP/sch/ds.t.sch1
      #去除表格或欄位編號含有$符號的
      sed -e '/\$/d'          $TOP/sch/ds.t.sch1 > $TOP/sch/ds.t.sch2
      sed -e '/^test_/d'      $TOP/sch/ds.t.sch2 > $TOP/sch/ds.t.sch1
      sed -e '/^temp_/d'      $TOP/sch/ds.t.sch1 > $TOP/sch/ds.t.sch2
      sed -e '/^tmp_/d'       $TOP/sch/ds.t.sch2 > $TOP/sch/ds.t.sch1
      sed -e '/_pk\^/d'       $TOP/sch/ds.t.sch1 > $TOP/sch/ds.t.sch2
      sed -e '/^a[a-oq-z]/d'  $TOP/sch/ds.t.sch2 > $TOP/sch/ds.t.sch3
      #處理完要停在 sch3 接下面

      rm -rf $TOP/sch/ds.sch
      rm -rf $TOP/sch/ds.42d
      rm -rf $TOP/sch/ds.t.sch1
      rm -rf $TOP/sch/ds.t.sch2
      mv $TOP/sch/ds.t.sch3 $TOP/sch/ds.sch
     
      if [ "$2" = "" ] ; then
         echo "INFO: 準備執行/ Preparing to execute reload_schfile.sh"
         $TOP/utl/bin/reload_gztz.sh 
         if [ $? -ne 0 ] ; then
            err="err"
         fi
      fi

      #echo "INFO: 準備畫面編譯用的ds.sch/ Preparing ds.sch for r.f ($TOP/sch/ext/ds.sch)"     #181225-00003#1  mark 
      #echo " "                                                                                #181225-00003#1  mark
      rm -rf $TOP/sch/ext/ds.sch
      rm -rf $TOP/sch/ext/ds.42d

      echo "ERPVER:"$ERPVER           #240919-00048 add
      if [ $ERPVER = "1.0" ] ; then   #240919-00048 add
         cat $TOP/sch/ds.sch | sed 's/\^10\^3594\^/^7^4^/g' > $TOP/sch/ext/ds.sch
         echo "INFO: 準備畫面編譯用的ds.sch/ Preparing ds.sch for r.f ($TOP/sch/ext/ds.sch)"      #181225-00003#1  add
      fi   #240919-00048 add
   fi
fi

if [ "$err" = "" ] ; then
   echo "INFO: 處理成功./ Success."
else
   echo "ERROR: 處理失敗.請查看錯誤訊息並依指示處理./ Unsuccess."
fi
cd $PW

