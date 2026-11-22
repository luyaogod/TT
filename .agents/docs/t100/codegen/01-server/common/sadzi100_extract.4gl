#+ Version..: T6-ERP-1.00.00 Build-00001
#+
#+ Filename.: sadzi100_extract
#+ Buildtype: p01樣板自動產生
#+ Memo.....: 解出資料庫內的指定樣板
#161230-00045 #1 2017/1/3 by jrg542 新增 多語言編輯畫面開窗時 title 說明
#180831-00006 #1 2018/08/31 by 00824 調整錯誤訊息顯示

IMPORT os
SCHEMA ds

GLOBALS "../../cfg/top_global.inc"

PUBLIC FUNCTION sadzi100_extract(lc_dzdy001,lc_dzdz002,ls_parm)

   DEFINE lc_dzdy001  LIKE dzdy_t.dzdy001   #樣板編號
   DEFINE lc_dzdy004  LIKE dzdy_t.dzdy004   #存放路徑
   DEFINE lc_dzdz002  LIKE dzdz_t.dzdz002   #版本
   DEFINE lc_dzdz003  LIKE dzdz_t.dzdz003   #CLOB欄位
   DEFINE ls_parm     STRING                #1: 寫回原來檔案路徑 2:寫到$TEMPDIR檔案路徑
   DEFINE ls_file     STRING
   DEFINE ls_cmd      STRING
   DEFINE ls_temp     STRING
   DEFINE li_chk      LIKE type_t.num5

   LOCATE lc_dzdz003 IN FILE

   SELECT dzdy004 INTO lc_dzdy004 FROM dzdy_t WHERE dzdy001 = lc_dzdy001

   #參數1 erp/mdl 
   IF ls_parm = "1" THEN  
      #定義檔案讀取路徑
      #IF lc_dzdy001[1,3] = "qry" OR lc_dzdy001[1,3] = "lng" THEN 
      IF lc_dzdy001[1,3] = "qry" OR lc_dzdy001[1,3] = "lng" OR lc_dzdy001[1,4] = "ntab" THEN  #17/01/03
         LET ls_file = lc_dzdy001,".4gl"
      ELSE 
         LET ls_file = lc_dzdy001,".template"
      END IF 
      
      IF lc_dzdy004[1,1] = "$" THEN
         LET ls_temp = lc_dzdy004 CLIPPED
         LET ls_cmd = ls_temp.subString(2,ls_temp.getIndexOf("/",1)-1)
         LET ls_file = os.Path.join(os.Path.join(FGL_GETENV(ls_cmd),ls_temp.subString(ls_temp.getIndexOf("/",1)+1,
                                                                 ls_temp.getLength())),ls_file)
      ELSE
         LET ls_file = os.Path.join(lc_dzdy004, ls_file)
      END IF 
   ELSE 
   
   #參數2 $TEMPDIR/code_i01.ver_NN
      LET ls_file = lc_dzdy001,".ver_",lc_dzdz002 USING "<<<<"
      LET ls_file = os.Path.join(FGL_GETENV("TEMPDIR"),ls_file)
      IF os.Path.exists(ls_file) THEN 
         IF os.Path.delete(ls_file) THEN END IF   
      END IF 
   END IF 
   SELECT COUNT(1) INTO li_chk FROM dzdz_t 
    WHERE  dzdz001 = lc_dzdy001    #樣板 
     AND  dzdz002 = lc_dzdz002     #版次 
     AND  dzdz003 IS NOT NULL 

   DISPLAY "檢核樣板內容:",lc_dzdy001 ," 是否為空(0:空/1:不為空):",li_chk
   IF li_chk = 0 THEN 
      DISPLAY "ERROR(20):樣板 ",lc_dzdy001," 內容為空, 請查詢adzi100!"
      DISPLAY "INFO:程式產生程序中止!"
      EXIT PROGRAM
   END IF 
   
   #讀取檔案
   SELECT dzdz003 INTO lc_dzdz003 FROM dzdz_t WHERE dzdz001 = lc_dzdy001 AND dzdz002 = lc_dzdz002

   IF lc_dzdz003 IS NULL THEN 
      DISPLAY "ERROR(20):樣板 ",lc_dzdy001," 內容為空, 請查詢adzi100!"
      DISPLAY "INFO:程式產生程序中止!"
      EXIT PROGRAM 
   END IF 
   #寫出檔案
   CALL lc_dzdz003.writeFile(ls_file)

   FREE lc_dzdz003

END FUNCTION


#+讀取計算md5 
PUBLIC FUNCTION sadzi100_counting_md5(ps_file) 
   DEFINE lch_channel base.Channel  
   DEFINE ps_file     STRING 
   DEFINE ls_temp     STRING
   DEFINE ls_cmd      STRING 

   #計算 MD5
   LET lch_channel = base.Channel.create()
   LET ls_cmd = "md5sum ",ps_file
   CALL lch_channel.setDelimiter("\t")
   CALL lch_channel.openPipe(ls_cmd ,"r")
   WHILE lch_channel.read(ls_temp)
      #display 'read :',current year to fraction(3)
      EXIT WHILE
   END WHILE
   CALL lch_channel.close()
   LET ls_temp = ls_temp.subString(1,32)
   RETURN ls_temp
END FUNCTION


#檢查樣板MD5是否跟資料庫的MD5一致 by ka0132
FUNCTION sadzi100_md5_chk(ps_id)
   DEFINE ps_id         STRING     #樣板代碼
   DEFINE ls_file       STRING     #樣板路徑&名稱
   DEFINE ls_path       STRING     #樣板路徑
   DEFINE ls_md5_o      CHAR(100)  #舊的md5(資料庫)
   DEFINE ls_md5_n      CHAR(100)  #新的md5(檔案)
   DEFINE ls_type_id    CHAR(100)  #樣板名稱
   DEFINE ls_loc        STRING
   
   #檢查此樣板是否合法(透過比對資料庫MD5)
   LET ls_loc = FGL_GETENV("ORACLE_SID")
   LET ls_loc = ls_loc.toLowerCase()
   IF NOT (ls_loc = 't10dev' OR ls_loc = 't10dit') THEN
   
      #先取得樣板名稱
      LET ls_type_id = 'code_',ps_id,'.template'
   
      #取出樣板路徑
      LET ls_path = FGL_GETENV("T100TEMPLATEPATH")
      IF cl_null(ls_path) THEN
         LET ls_path = FGL_GETENV("ERP")
      END IF
      LET ls_path = os.Path.join(ls_path,"mdl")
      LET ls_file = os.Path.join(ls_path,ls_type_id)
      LET ls_file = ls_file.trim()
      
      #檢查是否存在
      IF os.Path.exists(ls_file) THEN
         #去掉template
         LET ls_type_id = 'code_',ps_id
         #計算 MD5
         CALL sadzi100_counting_md5(ls_file) RETURNING ls_md5_n
         SELECT dzdz004 INTO ls_md5_o FROM dzdz_t
          WHERE dzdz002 = (SELECT MAX(dzdz002) FROM dzdz_t
                            WHERE dzdz001 = ls_type_id)
            AND dzdz001 = ls_type_id
         IF ls_md5_o <> ls_md5_n THEN
            #DISPLAY "ERROR(18):樣板 code_",ps_id," 資訊尚未更新, 請查詢adzi100!"
            #EXIT PROGRAM
         #  DISPLAY "WARNING(18):樣板code_",ps_id,"資訊尚未更新, 請查詢adzi100!"   #180831-00006
            DISPLAY "WARNING(18):樣板code_",ps_id,"實體檔與adzi100設定不同!"   #180831-00006
         ELSE
            DISPLAY "MD5比對成功, 樣板運作正常!"
         END IF
      ELSE
         DISPLAY "WARNING(19):樣板code_",ps_id,"不存在, 請查詢adzi100!"
         #DISPLAY "ERROR(19):樣板 code_",ps_id," 不存在, 請查詢adzi100!"
         #EXIT PROGRAM
      END IF
   END IF
   
END FUNCTION 

#檢查樣板MD5/產生器版本是否符合 by ka0132
FUNCTION sadzi100_template_chk(ps_id,pi_ver)
   DEFINE ps_id         STRING               #樣板代碼
   DEFINE pi_ver        LIKE type_t.num5     #產生器版本
   DEFINE li_ver        LIKE type_t.num5     #樣板對應的產生器版本
   DEFINE ls_file       STRING               #樣板路徑&名稱
   DEFINE ls_path       STRING               #樣板路徑
   DEFINE ls_md5_o      CHAR(100)            #舊的md5(資料庫)
   DEFINE ls_md5_n      CHAR(100)            #新的md5(檔案)
   DEFINE ls_type_id    CHAR(100)            #樣板名稱
   DEFINE ls_loc        STRING
   
   #檢查此樣板是否合法(透過比對資料庫MD5)
   LET ls_loc = FGL_GETENV("ORACLE_SID")
   LET ls_loc = ls_loc.toLowerCase()
   IF ls_loc = 'topprd' THEN #僅限制topprd
   
      #先取得樣板名稱
      LET ls_type_id = 'code_',ps_id,'.template'
   
      #取出樣板路徑
      LET ls_path = FGL_GETENV("T100TEMPLATEPATH")
      IF cl_null(ls_path) THEN
         LET ls_path = FGL_GETENV("ERP")
      END IF
      LET ls_path = os.Path.join(ls_path,"mdl")
      LET ls_file = os.Path.join(ls_path,ls_type_id)
      LET ls_file = ls_file.trim()
      
      #檢查是否存在
      IF os.Path.exists(ls_file) THEN
         #去掉template
         LET ls_type_id = 'code_',ps_id
         #計算 MD5
         CALL sadzi100_counting_md5(ls_file) RETURNING ls_md5_n
         SELECT dzdz004,dzdz010 INTO ls_md5_o,li_ver 
           FROM dzdz_t
          WHERE dzdz002 = (SELECT MAX(dzdz002) FROM dzdz_t
                            WHERE dzdz001 = ls_type_id)
            AND dzdz001 = ls_type_id
         
         #MD5不一致
         IF ls_md5_o <> ls_md5_n THEN
            DISPLAY "ERROR(18):樣板 code_",ps_id," 資訊尚未更新, 請查詢adzi100!"
            DISPLAY "INFO:程式產生程序中止!"
            EXIT PROGRAM
         ELSE
            DISPLAY "MD5比對成功, 樣板運作正常!"
         END IF
        
         #無版本資訊, 寫入當下產生器版本
         IF li_ver = 0 OR li_ver IS NULL THEN
            UPDATE dzdz_t SET dzdz010 = pi_ver
             WHERE dzdz002 = (SELECT MAX(dzdz002) FROM dzdz_t
                               WHERE dzdz001 = ls_type_id)
               AND dzdz001 = ls_type_id
         END IF
    
         #樣板版本過新, 無法套用
         IF li_ver > pi_ver THEN
            DISPLAY "ERROR(19):樣板 code_",ps_id," 版本過新, 請更新程式產生工具!"
            DISPLAY "INFO:程式產生程序中止!"
            EXIT PROGRAM
         END IF
      ELSE
         DISPLAY "ERROR(19):樣板 code_",ps_id," 不存在, 請查詢adzi100!"
         DISPLAY "INFO:程式產生程序中止!"
         EXIT PROGRAM
      END IF
   END IF
   
END FUNCTION 

