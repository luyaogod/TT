IMPORT os
SCHEMA ds

GLOBALS "../../cfg/top_global.inc"
 
GLOBALS
   DEFINE g_mask   om.SaxAttributes
   DEFINE gdnode_all     om.DomNode
END GLOBALS

DEFINE g_mask            om.SaxAttributes
DEFINE gnl_selted        om.NodeList

#browser資訊
DEFINE g_browser RECORD
       id    STRING,
       field STRING,
       var   STRING
       END RECORD
       
#單頭資訊
DEFINE g_master RECORD
       id    STRING,
       field STRING,
       var   STRING
       END RECORD
       
#單身資訊(by page)
DEFINE g_detail DYNAMIC ARRAY OF RECORD
       field STRING,
       var   STRING
       END RECORD

#單身資訊(by table)
DEFINE g_detail_by_tbl DYNAMIC ARRAY OF RECORD
       id    STRING,
       field STRING,
       var   STRING
       END RECORD

DEFINE gs_mask_fields STRING
DEFINE gs_gzyd002     LIKE gzyd_t.gzyd002
DEFINE gs_mdl         STRING
DEFINE gs_table_list  STRING
       
MAIN
   DEFINE lddoc_all      om.DomDocument  #tab資訊
   DEFINE ls_mod         STRING          #模組別
   DEFINE ls_prog        STRING          #程式名稱
   DEFINE ls_path        STRING          #相對路徑
   DEFINE ls_tab_path    STRING          #tab路徑
   DEFINE ls_compile     STRING          #執行編譯
   DEFINE ls_gzza001     LIKE gzza_t.gzza001
   DEFINE ls_app         LIKE gzza_t.gzza001
   DEFINE ls_app_s       STRING
   DEFINE ls_tmp         STRING
   DEFINE li_tmp         LIKE type_t.num10
   
   CALL cl_db_connect("ds",FALSE)
   
   #參數1(模組別)
   LET ls_mod = ARG_VAL(2)
   
   #參數2(程式名稱)
   LET ls_prog = ARG_VAL(3)
   LET gs_gzyd002 = ls_prog
   
   #參數3(不執行r.c)
   LET ls_compile = ARG_VAL(4)
   
   #如果輸入參數不完整則報錯
   IF cl_null(ls_mod) OR cl_null(ls_prog) THEN
      DISPLAY "ERROR:輸入資訊不完整, 請重新執行! "
      DISPLAY "      輸入參數1(",ls_mod ,")應為模組名稱! "
      DISPLAY "      輸入參數2(",ls_prog,")應為程式名稱! "
   END IF

   #c開頭為客製模組, 其餘為一般模組
   IF ls_mod = 'lib'  OR 
      ls_mod = 'sub'  OR 
      ls_mod = 'wss'  OR 
      ls_mod = 'qry'  OR 
      ls_mod = 'clib' OR 
      ls_mod = 'csub' OR 
      ls_mod = 'cwss' OR 
      ls_mod = 'cqry' THEN
      LET ls_path = "COM"
   ELSE      
      LET ls_path = "ERP"
   END IF
   
   #tabx檔更正為tab檔
   LET ls_tab_path = os.Path.join(os.Path.join(FGL_GETENV(ls_path),ls_mod),"dzx")
   LET ls_tab_path = os.Path.join(ls_tab_path,"tab")
   LET ls_tab_path = os.Path.join(ls_tab_path,ls_prog||".tab")
   
   #定義轉換用SaxAttributes 
   LET g_mask = om.SaxAttributes.create()
   CALL g_mask.addAttribute("general_module",ls_mod)  #模組名稱
   CALL g_mask.addAttribute("general_prefix",ls_prog) #程式名稱
   
   LET lddoc_all = om.DomDocument.createFromXmlFile(ls_tab_path)
   
   #若讀取不到資料代表tabx不存在
   TRY
      LET gdnode_all = lddoc_all.getDocumentElement() 
   CATCH
      DISPLAY "Info:無法讀取",ls_tab_path,", 遮罩產生工具結束!"
      EXIT PROGRAM
   END TRY
    
   LET gdnode_all = lddoc_all.getDocumentElement() 
   
   #取出使用的樣板
   CALL adzp155_get_mdl()
      
   
   #20160608
   #判斷是否為行業別程式, 再決定general_prefix
   #app_id為實際程式名稱(包含行業別), EX:azzi900_ic
   #general_prefix為排除行業別後的程式名稱, EX:azzi900
   #程式中除了section名稱外皆使用general_prefix,
   #目的是確保程式在引用時能夠取到正確的原程式段落
   LET ls_tmp = gs_mdl
   IF ls_tmp.getIndexOf('c',1) = 0 AND ls_tmp <> 'p00' THEN
      LET ls_app_s = ls_prog
      IF ls_app_s.getIndexOf('_',1) > 0 THEN
         LET ls_app = ls_app_s.subString(1,ls_app_s.getIndexOf('_',1)-1)
         LET li_tmp = 0
         SELECT COUNT(*) INTO li_tmp FROM gzza_t 
          WHERE gzza001 = ls_app
         IF li_tmp > 0 THEN
            DISPLAY "Info:此為行業別程式,mask定名為",ls_app
            CALL g_mask.addAttribute("general_prefix",ls_app)
         END IF
      END IF 
   END IF
      
   #取出有設遮罩的欄位  
   CALL adzp155_get_mask_fields()
      
   #取出reference對應表
   CALL adzp155_get_reference()
   
   #取出table資訊
   CALL adzp155_get_table()
   
   #組合欄位/變數資訊
   CALL adzp155_get_field_and_var()
   
   #根據table切分, 組合遮罩段落
   CALL adzp155_mask_by_table()
   
   #將資料寫入檔案(com/inc/erp/模組別/程式名稱_mask)
   CALL adzp155_write_file(ls_mod,ls_prog)
   
   #編譯產出檔案
   IF ls_compile <> 'n' OR cl_null(ls_compile) THEN
      CALL adzp155_compile(ls_mod,ls_prog)
   END IF
      
   #顯示完成
   DISPLAY 'Info:遮罩檔案(',ls_prog,'_mask.4gl)已產生完成!'
	  
END MAIN 

#+ 找出有設遮罩的欄位清單
PRIVATE FUNCTION adzp155_get_mdl()
   DEFINE ldnode_assembly    om.DomNode
   DEFINE ls_mdls            STRING
   
   LET gnl_selted      = gdnode_all.selectByTagName("assembly")
   LET ldnode_assembly = gnl_selted.item(1) 
   LET gs_mdl          = ldnode_assembly.getAttribute("type")

  LET ls_mdls = "i01,i02,i03,i04,i05,i06,i07,i08,i09,i10,i12,i13,t01,t02,c01a,c01b,c01c,c02a,c02b,c02c,c03a,c03b,c03c,q01,q02,q03,q04,q05,c04a"

    
   IF ls_mdls.getIndexOf(gs_mdl,1) = 0 THEN
      DISPLAY 'Info:該樣板不支援遮罩功能, 結束此程式!'
      EXIT PROGRAM
   END IF
   
END FUNCTION

#+ 找出有設遮罩的欄位清單
PRIVATE FUNCTION adzp155_get_mask_fields()
   DEFINE ls_sql     STRING
   DEFINE ls_gzyd003 LIKE gzyd_t.gzyd003
   DEFINE ls_gzzz001 LIKE gzzz_t.gzzz001
   DEFINE ls_wc      STRING
   
   #先根據該作業找出對應的主程式(1作業->1程式)
   SELECT gzzz002 FROM gzzz_t WHERE gzzz001 = gs_gzyd002
   
   #再根據此主程式找出所有對應的作業(1程式->N作業)
   LET ls_sql = ' SELECT gzzz001 FROM gzzz_t WHERE gzzz002 = ? '
   PREPARE prog_pre FROM ls_sql
   DECLARE prog_cur CURSOR FOR prog_pre
   OPEN prog_cur USING gs_gzyd002
   FOREACH prog_cur INTO ls_gzzz001
      LET ls_wc = ls_wc, "'", ls_gzzz001, "',"
   END FOREACH
   LET ls_wc = ls_wc.subString(1,ls_wc.getLength()-1)
   
   #主程式無設定, 判定為子程式, 直接套用該程式名稱
   IF cl_null(ls_wc) THEN
       LET ls_wc = "'",gs_gzyd002,"'"
   END IF
   
   #撈取有設遮罩的欄位, 此處不區分使用者身分與企業代碼
   LET ls_sql = " SELECT gzyd003 FROM gzyd_t WHERE gzyd002 IN (",ls_wc,") and gzyd004 = '2' "
   PREPARE sql_pre FROM ls_sql
   DECLARE sql_cur CURSOR FOR sql_pre

   #將撈取出有設定遮罩的欄位組合成字串清單
   FOREACH sql_cur INTO ls_gzyd003
      LET gs_mask_fields = gs_mask_fields, ls_gzyd003, ','
   END FOREACH

END FUNCTION

#+ 找出控件對應的實際欄位名稱
PRIVATE FUNCTION adzp155_get_reference()
   DEFINE ldnode_ref    om.DomNode
   DEFINE li_idx        INTEGER
   DEFINE ls_field      STRING #欄位名稱(呈現畫面上的控件名稱)
   DEFINE ls_component  STRING #欄位名稱(實際撈取的欄位名稱)
   DEFINE ls_sql        STRING #撈取的SQL
   
   #處理瀏覽頁reference
   LET gnl_selted = gdnode_all.selectByTagName("bs_reference")
   FOR li_idx = 1 TO gnl_selted.getLength()
      LET ldnode_ref = gnl_selted.item(li_idx)
      #browser欄位特別處理
      LET ls_component  = 'b_',ldnode_ref.getAttribute("field") 
      LET ls_sql        = ldnode_ref.getAttribute("ref_sql") 
      LET ls_field      = ls_sql.subString(ls_sql.getIndexOf('SELECT',1)+7,ls_sql.getIndexOf('FROM',1)-1)
      CALL g_mask.addAttribute(ls_component,ls_field)
   END FOR
   
   #處理一般單頭/身refernece
   LET gnl_selted = gdnode_all.selectByTagName("reference")
   FOR li_idx = 1 TO gnl_selted.getLength()
      LET ldnode_ref = gnl_selted.item(li_idx) 
      LET ls_component  = ldnode_ref.getAttribute("field") 
      LET ls_sql        = ldnode_ref.getAttribute("ref_sql") 
      LET ls_field      = ls_sql.subString(ls_sql.getIndexOf('SELECT',1)+7,ls_sql.getIndexOf('FROM',1)-1)
      CALL g_mask.addAttribute(ls_component,ls_field)
   END FOR
   
END FUNCTION

#+ 找出表格資訊(單頭/身)
PRIVATE FUNCTION adzp155_get_table()
   DEFINE ldnode_table    om.DomNode
   DEFINE ls_tbl_name     STRING
   DEFINE ls_tbl_pre      STRING

   #取出第一個單頭表格名稱(for 變數名稱用)
   LET gnl_selted = gdnode_all.selectByTagName("head")
   IF gnl_selted.getLength() > 0 THEN
      LET ldnode_table  = gnl_selted.item(1) 
      LET ls_tbl_name   = ldnode_table.getAttribute("id")
      CALL g_mask.addAttribute("master_tbl_name",ls_tbl_name)
      LET ls_tbl_pre    = ls_tbl_name.subString(1,ls_tbl_name.getIndexOf('_',1)-1)
      CALL g_mask.addAttribute("master_tbl_pre",ls_tbl_pre)
   END IF
   
   #取出第一個單身表格名稱(for 變數名稱用)
   LET gnl_selted = gdnode_all.selectByTagName("body")
   IF gnl_selted.getLength() > 0 THEN
      LET ldnode_table  = gnl_selted.item(1) 
      LET ls_tbl_name   = ldnode_table.getAttribute("id")
      CALL g_mask.addAttribute("detail_tbl_name",ls_tbl_name)
      LET ls_tbl_pre    = ls_tbl_name.subString(1,ls_tbl_name.getIndexOf('_',1)-1)
      CALL g_mask.addAttribute("detail_tbl_pre",ls_tbl_pre)
   END IF
    
END FUNCTION

#+ 取得所需的欄位/變數資訊
PRIVATE FUNCTION adzp155_get_field_and_var()
   DEFINE ldnode_field    om.DomNode
   DEFINE ls_tbl          STRING
   DEFINE ls_loc          STRING
   DEFINE ls_page         STRING
   DEFINE ls_page_id      STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_field_list   STRING
   DEFINE ls_var_list     STRING
   DEFINE ls_field        STRING
   DEFINE ls_var          STRING
   DEFINE ls_map_field    STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_title_pre    STRING
   
   #找出所有出現在畫面上的欄位(含browser,master,detail)
   LET gnl_selted = gdnode_all.selectByTagName("section") 
   
   FOR li_idx = 1 TO gnl_selted.getLength()
      LET ldnode_field  = gnl_selted.item(li_idx)
      IF ldnode_field.getAttribute("id") = 'global_var' THEN
	     LET ldnode_field  = gnl_selted.item(li_idx)
		 LET ldnode_field  = ldnode_field.getFirstChild()
         EXIT FOR
      END IF	  
   END FOR
   
   WHILE TRUE
      LET ls_loc        = ldnode_field.getAttribute("id")      #head/body/bs_field
      LET ls_page       = ldnode_field.getAttribute("page")    #組合變數名稱用
      LET ls_page_id    = ldnode_field.getAttribute("page_id") #組合變數名稱用
      LET ls_field_list = ldnode_field.getAttribute("value")
    
      #組合變數, 根據位置組出不同得變數名稱
      CASE ls_loc
         WHEN 'head'
            LET ls_title_pre = 'g_',g_mask.getValue('master_tbl_pre'),'_m.'
         WHEN 'body' 
            LET ls_title_pre = 'g_',adzp155_create_name(ls_page_id, g_mask.getValue('detail_tbl_pre'), "<<<"),'_d[l_ac].'
         WHEN 'bs_field'
            LET ls_title_pre = 'g_browser[g_cnt].b_'
      END CASE 
    
      #組合實際對應的欄位清單
      LET lst_token = base.StringTokenizer.create(ls_field_list, ',')
      LET ls_field_list = ''
      LET ls_var_list   = ''
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         #先去除()的部分
         IF ls_token.getIndexOf('(',1) THEN
            LET ls_token = ls_token.subString(1,ls_token.getIndexOf('(',1)-1)
         END IF
    
         #尋找是否為reference, 如果是則取出對應的名稱
         LET ls_map_field = ""
         IF ls_loc = 'bs_field' THEN
            LET ls_field = 'b_',ls_token
         ELSE
            LET ls_field = ls_token
         END IF

         LET ls_map_field = g_mask.getValue(ls_field)
         IF NOT cl_null(ls_map_field) THEN
            LET ls_field = ls_map_field
         ELSE
            #還原b_
            IF ls_loc = 'bs_field' THEN
               LET ls_field = ls_token
            END IF
         END IF
         
         #組合變數
         LET ls_var = ls_title_pre, ls_token
         LET ls_field = ls_field.trim(), ','
         LET ls_var   = ls_var, ','
         
         #判斷是否有設定遮罩並重組
         IF gs_mask_fields.getIndexOf(ls_field,1) THEN
            LET ls_field_list = ls_field_list, ls_field
            LET ls_var_list   = ls_var_list,   ls_var
         END IF
      END WHILE
    
      #記錄相關資訊到對應變數
      CASE ls_loc
         WHEN 'head'
            LET g_master.id    = g_mask.getValue('master_tbl_name')
            LET g_master.field = ls_field_list
            LET g_master.var   = ls_var_list
         WHEN 'body'
            LET g_detail[ls_page].field = ls_field_list
            LET g_detail[ls_page].var   = ls_var_list
         WHEN 'bs_field'
            LET g_browser.id    = 'browser'
            LET g_browser.field = ls_field_list
            LET g_browser.var   = ls_var_list
      END CASE 
	  
	  LET ldnode_field = ldnode_field.getNext()
	  
	  IF ldnode_field IS NULL THEN
	     EXIT WHILE
	  END IF 
	  
   END WHILE
   
END FUNCTION

#+ 取出各表格的遮罩欄位與變數名稱
PRIVATE FUNCTION adzp155_mask_by_table()
   DEFINE ldnode_table    om.DomNode
   DEFINE ls_tbl_name     STRING
   DEFINE ls_tbl_pre      STRING
   DEFINE li_idx          INTEGER
   DEFINE ls_pages        STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE ls_i07_field    STRING
   DEFINE ls_i07_var      STRING
   
   #先判定是否有browser
   IF g_browser.id IS NOT NULL THEN
      CALL adzp155_make_mask(g_browser.id,g_browser.field,g_browser.var)
   END IF
   
   #先判定是否有單頭
   IF g_master.id IS NOT NULL THEN
      #i07特別處理
      IF g_mask.getValue("master_tbl_pre") = g_mask.getValue("detail_tbl_pre") THEN
         IF NOT cl_null(g_master.field) THEN
            LET ls_i07_field = g_master.field,','
            LET ls_i07_var   = g_master.var  ,','
         END IF
      ELSE
         CALL adzp155_make_mask(g_master.id,g_master.field,g_master.var)
      END IF
   END IF
   
   #取出第一個單身表格名稱
   LET gnl_selted    = gdnode_all.selectByTagName("body")
   FOR li_idx = 1 TO gnl_selted.getLength()
      LET ldnode_table = gnl_selted.item(li_idx) 
      LET ls_tbl_name  = ldnode_table.getAttribute("id")
      LET ls_pages     = ldnode_table.getAttribute("page")

      IF gs_table_list.getIndexOf(ls_tbl_name||',',1) THEN
         DISPLAY '重複定義表格,即將略過',ls_tbl_name,'!'
         CONTINUE FOR
      END IF
      
      LET g_detail_by_tbl[li_idx].id = ls_tbl_name
      LET gs_table_list = gs_table_list,ls_tbl_name,','     
 
      #組合實際對應的欄位清單
      LET lst_token = base.StringTokenizer.create(ls_pages, ',')
      WHILE lst_token.hasMoreTokens()
         LET ls_token = lst_token.nextToken()
         LET g_detail_by_tbl[li_idx].field = g_detail_by_tbl[li_idx].field, g_detail[ls_token].field
         LET g_detail_by_tbl[li_idx].var   = g_detail_by_tbl[li_idx].var,   g_detail[ls_token].var
      END WHILE
      
      #組合單頭的欄位(i07 only)
      IF li_idx = 1 THEN
         LET g_detail_by_tbl[li_idx].field = ls_i07_field, g_detail_by_tbl[li_idx].field 
         LET g_detail_by_tbl[li_idx].var   = ls_i07_var,   g_detail_by_tbl[li_idx].var 
      END IF
      
      CALL adzp155_make_mask(g_detail_by_tbl[li_idx].id,g_detail_by_tbl[li_idx].field,g_detail_by_tbl[li_idx].var)
   END FOR
   
END FUNCTION

#+ 產生遮罩相關處理段
PRIVATE FUNCTION adzp155_make_mask(ps_tbl,ps_fields,ps_vars)
   DEFINE ps_tbl          STRING #table名稱
   DEFINE ps_fields       STRING #欄位名稱清單
   DEFINE ps_vars         STRING #變數名稱清單
   DEFINE ls_name         STRING
   DEFINE ls_value        STRING
   DEFINE lst_token       base.StringTokenizer
   DEFINE ls_token        STRING
   DEFINE lst_token2      base.StringTokenizer
   DEFINE ls_token2       STRING
   DEFINE li_idx         INTEGER
   DEFINE li_loc         INTEGER
   
   #mdl_tbl_name
   LET ls_name  = 'mdl_tbl_name'
   LET ls_value = ps_tbl
   CALL g_mask.addAttribute(ls_name,ls_value)
   
   #mdl_mark
   LET ls_name  = 'mdl_mark'
   IF cl_null(ps_fields) THEN
      LET ls_value = '#'
   ELSE
      LET ls_value = ''
   END IF
   CALL g_mask.addAttribute(ls_name,ls_value)
   
   LET li_idx = 0
   LET lst_token  = base.StringTokenizer.create(ps_fields, ',')
   LET lst_token2 = base.StringTokenizer.create(ps_vars  , ',')
   WHILE lst_token.hasMoreTokens()
      LET ls_token  = lst_token.nextToken()
      LET ls_token2 = lst_token2.nextToken()
      
      #排除別名
      IF ls_token.getIndexOf('.',1) > 0 THEN
         LET ls_token = ls_token.subString(ls_token.getIndexOf('.',1)+1,ls_token.getLength())
      END IF
      
      #排除非實際欄位
      IF ls_token.getIndexOf('_desc',1) > 0 THEN
         CONTINUE WHILE
      END IF
      
      #統一使用l_ac
      #IF ls_token2.getIndexOf('g_cnt',1) > 0 THEN
      #   LET ls_token2 = cl_replace_str(ls_token2,'g_cnt','l_ac')
      #END IF

      LET li_idx = li_idx + 1
      
      CASE
         #單身
         WHEN ls_token2.getIndexOf('[',1) > 0 
            LET li_loc = ls_token2.getIndexOf('[',1)
         #單頭
         WHEN ls_token2.getIndexOf('.',1) > 0 
            LET li_loc = ls_token2.getIndexOf('.',1)
      END CASE
      
      #mdl_field(1~N)
      LET ls_name  = adzp155_create_name(li_idx, "mdl_field", "<<<")
      LET ls_value = ls_token
      CALL g_mask.addAttribute(ls_name,ls_value)
      
      #mdl_var(1~N)
      LET ls_name  = adzp155_create_name(li_idx, "mdl_var", "<<<")
      LET ls_value = ls_token2
      CALL g_mask.addAttribute(ls_name,ls_value)
      
      #mdl_var_o(1~N)
      LET ls_name  = adzp155_create_name(li_idx, "mdl_var_o", "<<<")
      LET ls_value = ls_token2.subString(1,li_loc-1), 
                     '_mask_o',
                     ls_token2.subString(li_loc,ls_token2.getLength()) 
      CALL g_mask.addAttribute(ls_name,ls_value)
      
      #mdl_var_n(1~N)
      LET ls_name  = adzp155_create_name(li_idx, "mdl_var_n", "<<<")
      LET ls_value = ls_token2.subString(1,li_loc-1), 
                     '_mask_n',
                     ls_token2.subString(li_loc,ls_token2.getLength()) 
      CALL g_mask.addAttribute(ls_name,ls_value)
   END WHILE
   
   CALL g_mask.addAttribute("mdl_mdls",li_idx)
   
   LET ls_name  = "general_mask_funcs"
   LET ls_value = g_mask.getValue(ls_name)
   LET ls_value = ls_value,'\n',
                  adzp155_make_slice("a64")
   
   #browser無法修改/不須備份
   IF NOT (ps_tbl = 'browser' OR gs_mdl.getCharAt(1) = 'q') THEN   
      LET ls_value = ls_value,'\n',
                     adzp155_make_slice("a65")
   END IF
   CALL g_mask.addAttribute(ls_name,ls_value)
   
END FUNCTION

#+ 讀取片段樣板並進行取代(欄位檢查段落)
PRIVATE FUNCTION adzp155_make_slice(ps_mdl)
   DEFINE ps_mdl          STRING
   DEFINE ls_mdl          STRING
   DEFINE ls_read         STRING
   DEFINE ls_return       STRING
   DEFINE ls_return_tmp   STRING
   DEFINE ls_text         STRING
   DEFINE lchannel_read   base.Channel
   DEFINE lc_write        base.Channel #回傳用
   DEFINE ls_mdlpath      STRING
   
   #先讀取T100TEMPLATEPATH, 若無則採用$ERP/mdl
   LET ls_mdlpath = FGL_GETENV("T100TEMPLATEPATH")
   IF cl_null(ls_mdlpath) THEN
      LET ls_mdlpath = FGL_GETENV("ERP")
      LET ls_mdlpath = os.Path.join(ls_mdlpath,"mdl")
   END IF
   LET ls_mdlpath = os.Path.join(ls_mdlpath,"slice")
      
   #定義取用樣板檔案
   LET ls_mdl = "code_",ps_mdl,".template"
   LET ls_mdl = os.Path.join(ls_mdlpath,ls_mdl)

   #開啟樣板檔
   LET lchannel_read = base.Channel.create()
   CALL lchannel_read.setDelimiter("")
   CALL lchannel_read.openFile( ls_mdl CLIPPED, "r" )
   
   WHILE NOT lchannel_read.isEof()
   
      LET ls_text = ""
      LET ls_read = lchannel_read.readLine()
      
      LET ls_return_tmp = ""
      
      #產生code部分
      CASE
         #slice專用迴圈1
         WHEN ls_read.getIndexOf("#mdls - Start -",1)
            CALL adzp155_make_mdls(lchannel_read,1) RETURNING lchannel_read,ls_return_tmp
            LET ls_read = ""
         
         #slice專用迴圈2
         WHEN ls_read.getIndexOf("#mdls2 - Start -",1)
            CALL adzp155_make_mdls(lchannel_read,2) RETURNING lchannel_read,ls_return_tmp
            LET ls_read = ""         
         
         #slice專用迴圈3
         WHEN ls_read.getIndexOf("#mdls3 - Start -",1)
            CALL adzp155_make_mdls(lchannel_read,3) RETURNING lchannel_read,ls_return_tmp
            LET ls_read = ""
            
      END CASE
      
      LET ls_return = ls_return, ls_return_tmp

      #行代換/ 對 ${} 置換
      IF ls_read.getIndexOf("${",1) AND 
         ( ls_read.getIndexOf("${",1) < ls_read.getIndexOf("}",1) ) THEN
         LET ls_text = adzp155_line_replace(ls_read)
      ELSE
         LET ls_text = ls_read
      END IF
      LET ls_return = ls_return, ls_text, '\n'
   END WHILE  
   
   CALL lchannel_read.close()
   RETURN ls_return

END FUNCTION

#+ mdl產生
PRIVATE FUNCTION adzp155_make_mdls(ps_read,pi_num)
   DEFINE ps_read             base.Channel
   DEFINE pi_num              LIKE type_t.num5 
   DEFINE ls_mdl              DYNAMIC ARRAY OF STRING #第幾組mdl
   DEFINE li_mdl_num          LIKE type_t.num5 
   DEFINE li_cnt              LIKE type_t.num5 
   DEFINE ls_is               STRING 
   DEFINE ls_return           STRING
   DEFINE ls_tmp              STRING
   
   #先取出mdl段落之樣版
   WHILE TRUE 
      LET ls_tmp = ps_read.readLine()
      IF ls_tmp.getIndexOf("#mdls -  End  -",1) THEN
         LET ls_mdl[1] = ls_mdl[1].subString(1,ls_mdl[1].getLength()-1)
         EXIT WHILE 
      END IF
      LET ls_mdl[1] = ls_mdl[1], ls_tmp, '\n'
   END WHILE 
   
   #總數量
   LET ls_tmp = adzp155_create_name(pi_num, "mdl_mdls", "<<<") 
   LET li_mdl_num = g_mask.getValue(ls_tmp)
   
   #根據mdl數量產生對應的mdl段落並進行資料取代
   FOR li_cnt = 2 TO li_mdl_num
      
      #先將${mdls}取代掉
      LET ls_is = li_cnt
      LET ls_mdl[li_cnt] = adzp155_replace_mdls(ls_mdl[1],li_cnt)
                        
      #若資料未取代完成則重複返回
      WHILE ls_mdl[li_cnt].getIndexOf("${",1) > 0  #}
         LET ls_mdl[li_cnt] = adzp155_line_replace(ls_mdl[li_cnt])
      END WHILE  
      LET ls_return = ls_return, ls_mdl[li_cnt]
      
      IF li_cnt < li_mdl_num THEN
         LET ls_return = ls_return, '\n'
      END IF
      
   END FOR    
   
   RETURN ps_read, ls_return
   
END FUNCTION 

#+ 將${mdl}取代為號碼
PRIVATE FUNCTION adzp155_replace_mdls(ps_mdl, pi_i)
   DEFINE ps_mdl   STRING 
   DEFINE pi_i     STRING 
   DEFINE li_s     LIKE type_t.num5 
   DEFINE li_e     LIKE type_t.num5 
   DEFINE ls_mdl   STRING 

   WHILE TRUE 
      LET li_s = ps_mdl.getIndexOf("${mdl}",1) - 1
      LET li_e = ps_mdl.getLength()
      #取代${mdl}
      LET ls_mdl = ps_mdl.subString(1,li_s), pi_i, ps_mdl.subString((li_s+7),li_e)
      LET ps_mdl = ls_mdl
      IF ls_mdl.getIndexOf("${mdl}",1) < 1 THEN 
         EXIT WHILE
      END IF 
   END WHILE 
    
   RETURN ls_mdl
   
END FUNCTION 

#+ 取得變數名稱
PRIVATE FUNCTION adzp155_create_name(ps_idx,ps_name,ps_type)
   DEFINE ps_idx     STRING
   DEFINE ps_name    STRING
   DEFINE ps_type    STRING
   DEFINE ls_return  STRING

   IF ps_idx = "1" AND ps_type = "<<<" THEN
      LET ls_return = ps_name
   ELSE
      IF NOT cl_null(ps_idx USING ps_type) THEN
         LET ls_return = ps_name, ps_idx USING ps_type
      ELSE
         LET ls_return = ps_name, ps_idx
      END IF
   END IF

   RETURN ls_return

END FUNCTION

#+ 逐行代換
PRIVATE FUNCTION adzp155_line_replace(ls_read)
   DEFINE ls_read     STRING
   DEFINE ls_text     STRING
   DEFINE ls_tag      STRING
   DEFINE li_pos1     LIKE type_t.num10
   DEFINE li_pos2     LIKE type_t.num10
   DEFINE ls_temp     STRING                #暫存properties資料用

   LET li_pos1 = ls_read.getIndexOf("${",1)
   LET li_pos2 = ls_read.getIndexOf("}", li_pos1)

   IF li_pos1 > 0 AND li_pos2 > 0 AND li_pos1 < li_pos2 THEN
      LET ls_text = ""
      LET ls_tag = ls_read.subString(li_pos1 +2, li_pos2 -1 ) #取出要置換的tag

      #由SaxAttribute內取出值進行代換
      #不在行首
      IF li_pos1 > 1 THEN
         LET ls_text = ls_read.subString(1,li_pos1 -1)
      END IF

      #中間段
      LET ls_temp = g_mask.getValue(ls_tag) CLIPPED
         
      LET ls_text = ls_text,g_mask.getValue(ls_tag) CLIPPED,
                       ls_read.subString(li_pos2+1,ls_read.getLength())

      #遞迴處理同行其他組
      IF ls_text.getIndexOf("${",1) THEN #}
         LET ls_text = adzp155_line_replace(ls_text)
      END IF

   END IF

   RETURN ls_text

END FUNCTION

#+ 將產出段落回寫到檔案中
PRIVATE FUNCTION adzp155_write_file(ps_mod,ps_prog)
   DEFINE ps_mod           STRING
   DEFINE ps_prog          STRING
   DEFINE ls_file          STRING
   DEFINE ls_code_file     STRING
   DEFINE lchannel_write   base.Channel
   
   LET ls_file = g_mask.getValue('general_mask_funcs')

   #產出程式路徑
   LET ls_code_file = os.Path.join(FGL_GETENV('COM'),'inc')
   LET ls_code_file = os.Path.join(ls_code_file,'erp')
   LET ls_code_file = os.Path.join(ls_code_file,ps_mod)
   LET ls_code_file = os.Path.join(ls_code_file,ps_prog||"_mask.4gl")
   
   DISPLAY "\nInfo:遮罩產生工具開始處理:"
   
   #先行移除舊檔案
   IF os.Path.delete(ls_code_file) THEN
      DISPLAY "Info:刪除舊檔案:",ls_code_file
   END IF
   
   #判斷是否砍除成功
   IF NOT os.Path.exists(ls_code_file) THEN
      DISPLAY "Info:舊案刪除成功:",ls_code_file
   ELSE
      DISPLAY "ERROR:舊檔案刪除失敗:",ls_code_file
      EXIT PROGRAM
   END IF
   
   LET lchannel_write = base.Channel.create()
   CALL lchannel_write.setDelimiter("")
   
   TRY
      DISPLAY "Info:產生檔案位置:",ls_code_file
      CALL lchannel_write.openFile( ls_code_file, "w" )
      CALL lchannel_write.write(ls_file)
      CALL lchannel_write.close()
   CATCH
      DISPLAY "Info:產生檔案位置異常(",ls_code_file,")!"
      EXIT PROGRAM
   END TRY 
    
   DISPLAY "Info:程式產生完成!"
   
END FUNCTION

#+ 編譯程式
PRIVATE FUNCTION adzp155_compile(ps_mod,ps_prog)
   DEFINE ps_mod        STRING #模組別
   DEFINE ps_prog       STRING #程式名稱
   DEFINE ls_cmd        STRING #指令
   
   #切換到該模組的路徑下
   IF NOT os.Path.chdir(os.Path.join(FGL_GETENV(ps_mod.toUpperCase()),"4gl")) THEN
      DISPLAY "Warning:無法切換到指定路徑(",os.Path.join(FGL_GETENV(ps_mod.toUpperCase()),"4gl"),")"
   ELSE
      DISPLAY "Info:切換到指定路徑(",os.Path.join(FGL_GETENV(ps_mod.toUpperCase()),"4gl"),")"
   END IF
   
   #進行編譯
   LET ls_cmd = "r.c ",ps_prog.trim()
   DISPLAY 'Info:目錄:',os.Path.pwd()," 指令:",ls_cmd
   RUN ls_cmd

END FUNCTION

